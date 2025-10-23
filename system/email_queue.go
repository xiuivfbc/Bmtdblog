package system

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// EmailTask 邮件任务结构
type EmailTask struct {
	ID       string    `json:"id"`
	To       string    `json:"to"`
	Subject  string    `json:"subject"`
	Body     string    `json:"body"`
	Retry    int       `json:"retry"`
	MaxRetry int       `json:"max_retry"`
	CreateAt time.Time `json:"create_at"`
}

// EmailQueueStats 队列统计信息
type EmailQueueStats struct {
	QueueSize      int64 `json:"queue_size"`
	FailedSize     int64 `json:"failed_size"`
	WorkerCount    int   `json:"worker_count"`
	ProcessedTotal int64 `json:"processed_total"`
	FailedTotal    int64 `json:"failed_total"`
}

// EmailQueue 邮件队列管理器
type EmailQueue struct {
	redis      *RedisCacheClient
	queueKey   string
	failKey    string
	delayedKey string // 延迟队列key

	// 动态Worker池配置
	minWorkers      int                  // 最小Worker数量
	maxWorkers      int                  // 最大Worker数量
	currentWorkers  int                  // 当前Worker数量
	workers         map[int]*EmailWorker // 改为map管理Worker
	workerIDCounter int                  // Worker ID计数器

	// 扩缩容配置
	scaleUpThreshold   int64         // 扩容阈值：队列长度
	scaleDownThreshold int64         // 缩容阈值：队列长度
	idleTimeout        time.Duration // Worker空闲超时时间

	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	workerMutex sync.RWMutex // Worker管理锁
	stats       EmailQueueStats
	statsMutex  sync.RWMutex
	sendFunc    func(to, subject, body string) error // 邮件发送回调函数
}

// EmailWorker 邮件工作者
type EmailWorker struct {
	id         int
	queue      *EmailQueue
	ctx        context.Context
	cancel     context.CancelFunc
	lastActive time.Time    // 上次活动时间
	isRunning  bool         // 运行状态
	mutex      sync.RWMutex // 状态锁
}

// 全局邮件队列实例
var EmailQueueInstance *EmailQueue

// max 返回两个整数中的最大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// InitEmailQueue 初始化邮件队列
func InitEmailQueue(workerCount int) error {
	if Redis == nil || !Redis.IsAvailable() {
		Logger.Warn("Redis不可用，邮件队列将使用同步模式")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	EmailQueueInstance = &EmailQueue{
		redis:      Redis,
		queueKey:   "bmtdblog:email:queue",
		failKey:    "bmtdblog:email:failed",
		delayedKey: "bmtdblog:email:delayed", // 延迟队列key

		// 动态池配置
		minWorkers:      max(1, workerCount/2),      // 最小Worker数：传入值的一半
		maxWorkers:      workerCount * 3,            // 最大Worker数：传入值的3倍
		currentWorkers:  0,                          // 当前Worker数
		workers:         make(map[int]*EmailWorker), // Worker映射
		workerIDCounter: 0,                          // ID计数器

		// 扩缩容阈值配置
		scaleUpThreshold:   int64(workerCount * 10), // 队列长度 > 10*worker数时扩容
		scaleDownThreshold: int64(workerCount * 2),  // 队列长度 < 2*worker数时缩容
		idleTimeout:        5 * time.Minute,         // Worker空闲5分钟后可回收

		ctx:    ctx,
		cancel: cancel,
		stats: EmailQueueStats{
			WorkerCount: workerCount,
		},
		sendFunc: sendEmailSync, // 默认使用同步发送
	}

	// 🚀 启动初始的Worker数量（使用最小值）
	for i := 0; i < EmailQueueInstance.minWorkers; i++ {
		EmailQueueInstance.workerIDCounter++
		EmailQueueInstance.startWorker(EmailQueueInstance.workerIDCounter)
	}

	// 🚀 启动延迟任务处理器
	EmailQueueInstance.wg.Add(1)
	go EmailQueueInstance.processDelayedTasks()

	// 🚀 启动动态扩缩容监控器
	EmailQueueInstance.wg.Add(1)
	go EmailQueueInstance.monitorAndScale()

	Logger.Info("动态邮件队列已启动",
		"min_workers", EmailQueueInstance.minWorkers,
		"max_workers", EmailQueueInstance.maxWorkers,
		"current_workers", EmailQueueInstance.currentWorkers,
		"scale_up_threshold", EmailQueueInstance.scaleUpThreshold,
		"scale_down_threshold", EmailQueueInstance.scaleDownThreshold)
	return nil
}

// SetEmailSender 设置邮件发送函数
func SetEmailSender(sendFunc func(to, subject, body string) error) {
	if EmailQueueInstance != nil {
		EmailQueueInstance.sendFunc = sendFunc
	}
}

// startWorker 启动一个新的Worker
func (eq *EmailQueue) startWorker(workerID int) {
	eq.workerMutex.Lock()
	defer eq.workerMutex.Unlock()

	if len(eq.workers) >= eq.maxWorkers {
		Logger.Warn("已达到最大Worker数量，无法启动更多Worker",
			"current", len(eq.workers),
			"max", eq.maxWorkers)
		return
	}

	// 如果worker ID已存在，生成新的ID
	if _, exists := eq.workers[workerID]; exists {
		eq.workerIDCounter++
		workerID = eq.workerIDCounter
	}

	// 创建Worker专属的context
	workerCtx, workerCancel := context.WithCancel(eq.ctx)

	worker := &EmailWorker{
		id:         workerID,
		queue:      eq,
		ctx:        workerCtx,
		cancel:     workerCancel,
		lastActive: time.Now(),
		isRunning:  true,
	}

	eq.workers[workerID] = worker

	eq.wg.Add(1)
	go worker.Start()

	Logger.Info("启动新的EmailWorker",
		"worker_id", workerID,
		"current_workers", eq.currentWorkers)
}

// stopWorker 停止指定的Worker
func (eq *EmailQueue) stopWorker(workerID int) {
	eq.workerMutex.Lock()
	defer eq.workerMutex.Unlock()

	worker, exists := eq.workers[workerID]
	if !exists {
		Logger.Warn("Worker不存在", "worker_id", workerID)
		return
	}

	// 标记为停止状态
	worker.mutex.Lock()
	worker.isRunning = false
	worker.mutex.Unlock()

	// 取消Worker的context
	worker.cancel()

	// 从workers map中移除
	delete(eq.workers, workerID)

	Logger.Info("停止EmailWorker",
		"worker_id", workerID,
		"current_workers", len(eq.workers))
}

// PushEmailTask 推送邮件任务到队列
func PushEmailTask(to, subject, body string) error {
	if EmailQueueInstance == nil {
		// 队列不可用，使用默认同步发送
		return sendEmailSync(to, subject, body)
	}

	task := EmailTask{
		ID:       generateTaskID(),
		To:       to,
		Subject:  subject,
		Body:     body,
		Retry:    0,
		MaxRetry: 3,
		CreateAt: time.Now(),
	}

	return EmailQueueInstance.Push(task)
}

// Push 推送任务到队列
func (eq *EmailQueue) Push(task EmailTask) error {
	if eq.redis == nil || !eq.redis.IsAvailable() {
		// Redis不可用，降级为同步发送
		return eq.sendFunc(task.To, task.Subject, task.Body)
	}

	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化邮件任务失败: %v", err)
	}

	// 使用LPUSH将任务推入队列左端
	_, err = eq.redis.client.LPush(eq.ctx, eq.queueKey, taskJSON).Result()
	if err != nil {
		// Redis操作失败，降级为同步发送
		Logger.Error("推送邮件任务到队列失败，降级为同步发送", "err", err)
		return eq.sendFunc(task.To, task.Subject, task.Body)
	}

	Logger.Debug("邮件任务已推入队列", "task_id", task.ID, "to", task.To)
	return nil
}

// PushWithDelay 推送任务到延迟队列
func (eq *EmailQueue) PushWithDelay(task EmailTask, delaySeconds int) error {
	if eq.redis == nil || !eq.redis.IsAvailable() {
		// Redis不可用，降级为同步发送
		return eq.sendFunc(task.To, task.Subject, task.Body)
	}

	// 计算执行时间戳
	executeTime := time.Now().Add(time.Duration(delaySeconds) * time.Second).Unix()

	taskJSON, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("序列化延迟邮件任务失败: %v", err)
	}

	// 使用ZAdd将任务推入延迟队列，score为执行时间戳
	_, err = eq.redis.client.ZAdd(eq.ctx, eq.delayedKey, redis.Z{
		Score:  float64(executeTime),
		Member: taskJSON,
	}).Result()

	if err != nil {
		// Redis操作失败，降级为同步发送
		Logger.Error("推送延迟邮件任务到队列失败，降级为同步发送", "err", err)
		return eq.sendFunc(task.To, task.Subject, task.Body)
	}

	Logger.Debug("延迟邮件任务已推入队列",
		"task_id", task.ID,
		"to", task.To,
		"delay_seconds", delaySeconds,
		"execute_time", time.Unix(executeTime, 0))
	return nil
}

// Start 启动邮件工作者
func (ew *EmailWorker) Start() {
	defer ew.queue.wg.Done()

	Logger.Info("EmailWorker启动", "worker_id", ew.id)

	for {
		select {
		case <-ew.ctx.Done():
			Logger.Info("EmailWorker停止", "worker_id", ew.id)
			return
		default:
			if err := ew.ProcessTask(); err != nil {
				// 处理错误，避免worker崩溃
				Logger.Error("EmailWorker处理任务出错",
					"worker_id", ew.id,
					"err", err)
				time.Sleep(1 * time.Second) // 错误后稍微等待
			}
		}
	}
}

// ProcessTask 处理单个邮件任务
func (ew *EmailWorker) ProcessTask() error {
	// 更新Worker状态
	ew.mutex.Lock()
	ew.isRunning = true
	ew.lastActive = time.Now()
	ew.mutex.Unlock()

	// 使用BRPOP阻塞式获取任务（右端弹出）
	result, err := ew.queue.redis.client.BRPop(ew.ctx, 5*time.Second, ew.queue.queueKey).Result()
	if err != nil {
		// 任务完成，更新状态
		ew.mutex.Lock()
		ew.isRunning = false
		ew.lastActive = time.Now()
		ew.mutex.Unlock()

		if err.Error() == "redis: nil" {
			// 队列为空，继续等待
			return nil
		}
		return fmt.Errorf("从队列获取任务失败: %v", err)
	}

	if len(result) < 2 {
		// 任务完成，更新状态
		ew.mutex.Lock()
		ew.isRunning = false
		ew.lastActive = time.Now()
		ew.mutex.Unlock()
		return fmt.Errorf("无效的队列数据")
	}

	// 解析任务
	var task EmailTask
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		Logger.Error("反序列化邮件任务失败", "err", err, "data", result[1])
		// 任务完成，更新状态
		ew.mutex.Lock()
		ew.isRunning = false
		ew.lastActive = time.Now()
		ew.mutex.Unlock()
		return nil // 跳过无效任务
	}

	// 执行邮件发送
	Logger.Debug("EmailWorker开始处理任务",
		"worker_id", ew.id,
		"task_id", task.ID,
		"to", task.To)

	if err := ew.sendEmail(task); err != nil {
		result := ew.handleFailedTask(task, err)
		// 任务完成，更新状态
		ew.mutex.Lock()
		ew.isRunning = false
		ew.lastActive = time.Now()
		ew.mutex.Unlock()
		return result
	}

	// 任务完成，更新状态
	ew.mutex.Lock()
	ew.isRunning = false
	ew.lastActive = time.Now()
	ew.mutex.Unlock()

	Logger.Info("邮件发送成功",
		"worker_id", ew.id,
		"task_id", task.ID,
		"to", task.To)

	return nil
}

// sendEmail 发送邮件
func (ew *EmailWorker) sendEmail(task EmailTask) error {
	// 调用队列配置的邮件发送函数
	err := ew.queue.sendFunc(task.To, task.Subject, task.Body)
	if err == nil {
		ew.queue.incrementProcessedCount()
	} else {
		ew.queue.incrementFailedCount()
	}
	return err
}

// handleFailedTask 处理失败的任务
func (ew *EmailWorker) handleFailedTask(task EmailTask, err error) error {
	Logger.Error("邮件发送失败",
		"worker_id", ew.id,
		"task_id", task.ID,
		"to", task.To,
		"retry", task.Retry,
		"err", err)

	task.Retry++

	if task.Retry < task.MaxRetry {
		// 🚀 优化：使用延迟队列代替Sleep，不阻塞Worker
		// 延迟策略：30秒、60秒、90秒（递增）
		delaySeconds := task.Retry * 30
		Logger.Info("任务将延迟重试",
			"task_id", task.ID,
			"retry_count", task.Retry,
			"delay_seconds", delaySeconds)
		return ew.queue.PushWithDelay(task, delaySeconds)
	} else {
		// 达到最大重试次数，移入失败队列
		return ew.moveToFailedQueue(task, err)
	}
}

// moveToFailedQueue 移动到失败队列
func (ew *EmailWorker) moveToFailedQueue(task EmailTask, err error) error {
	failedTask := map[string]interface{}{
		"task":      task,
		"error":     err.Error(),
		"failed_at": time.Now(),
		"worker_id": ew.id,
	}

	failedJSON, jsonErr := json.Marshal(failedTask)
	if jsonErr != nil {
		Logger.Error("序列化失败任务出错", "err", jsonErr)
		return jsonErr
	}

	_, redisErr := ew.queue.redis.client.LPush(ew.ctx, ew.queue.failKey, failedJSON).Result()
	if redisErr != nil {
		Logger.Error("移动失败任务到失败队列出错", "err", redisErr)
		return redisErr
	}

	Logger.Warn("邮件任务已移入失败队列",
		"task_id", task.ID,
		"worker_id", ew.id,
		"to", task.To)

	return nil
}

// processDelayedTasks 处理延迟任务的后台处理器
func (eq *EmailQueue) processDelayedTasks() {
	defer eq.wg.Done()

	Logger.Info("延迟任务处理器启动")

	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-eq.ctx.Done():
			Logger.Info("延迟任务处理器停止")
			return
		case <-ticker.C:
			if err := eq.moveExpiredTasksToQueue(); err != nil {
				Logger.Error("处理延迟任务出错", "err", err)
			}
		}
	}
}

// moveExpiredTasksToQueue 将到期的延迟任务移动到正常队列
func (eq *EmailQueue) moveExpiredTasksToQueue() error {
	if eq.redis == nil || !eq.redis.IsAvailable() {
		return nil
	}

	now := time.Now().Unix()

	// 获取所有到期的任务（score <= now）
	result, err := eq.redis.client.ZRangeByScore(eq.ctx, eq.delayedKey, &redis.ZRangeBy{
		Min:    "0",
		Max:    fmt.Sprintf("%d", now),
		Offset: 0,
		Count:  100, // 每次最多处理100个任务
	}).Result()

	if err != nil {
		return fmt.Errorf("获取到期延迟任务失败: %v", err)
	}

	processedCount := 0
	for _, taskJSON := range result {
		// 原子操作：从延迟队列移除
		removed, err := eq.redis.client.ZRem(eq.ctx, eq.delayedKey, taskJSON).Result()
		if err != nil {
			Logger.Error("从延迟队列移除任务失败", "err", err, "task", taskJSON)
			continue
		}

		if removed == 0 {
			// 任务已被其他进程处理
			continue
		}

		// 解析任务
		var task EmailTask
		if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
			Logger.Error("反序列化延迟任务失败", "err", err, "task", taskJSON)
			continue
		}

		// 推入正常队列
		if err := eq.Push(task); err != nil {
			Logger.Error("将延迟任务推入正常队列失败", "err", err, "task_id", task.ID)
			// 如果推入失败，可以考虑重新放回延迟队列
			continue
		}

		processedCount++
		Logger.Debug("延迟任务已移入正常队列",
			"task_id", task.ID,
			"to", task.To,
			"original_delay", task.Retry*30)
	}

	if processedCount > 0 {
		Logger.Info("处理延迟任务完成", "processed_count", processedCount)
	}

	return nil
}

// monitorAndScale 监控队列压力并自动调整worker数量
func (eq *EmailQueue) monitorAndScale() {
	ticker := time.NewTicker(30 * time.Second) // 每30秒检查一次
	defer ticker.Stop()
	defer eq.wg.Done()

	for {
		select {
		case <-eq.ctx.Done():
			Logger.Info("自动扩缩容监控器停止")
			return
		case <-ticker.C:
			eq.checkAndScale()
		}
	}
}

// checkAndScale 检查队列状态并执行缩放
func (eq *EmailQueue) checkAndScale() {
	eq.workerMutex.Lock()
	defer eq.workerMutex.Unlock()

	// 获取当前队列长度
	queueLength, err := eq.redis.client.LLen(eq.ctx, eq.queueKey).Result()
	if err != nil {
		Logger.Error("获取队列长度失败", "err", err)
		return
	}

	Logger.Debug("队列状态检查",
		"queue_length", queueLength,
		"current_workers", len(eq.workers),
		"min_workers", eq.minWorkers,
		"max_workers", eq.maxWorkers)

	// 扩容条件：队列长度超过阈值且workers未达到最大值
	if queueLength > eq.scaleUpThreshold && len(eq.workers) < eq.maxWorkers {
		newWorkerID := eq.workerIDCounter + 1
		eq.workerIDCounter = newWorkerID

		eq.startWorker(newWorkerID)
		Logger.Info("自动扩容worker",
			"new_worker_id", newWorkerID,
			"total_workers", len(eq.workers),
			"queue_length", queueLength)
		return
	}

	// 缩容条件：队列长度低于阈值且workers超过最小值
	if queueLength < eq.scaleDownThreshold && len(eq.workers) > eq.minWorkers {
		// 找到最旧的空闲worker
		var oldestWorker *EmailWorker
		var oldestID int
		oldestTime := time.Now()

		for id, worker := range eq.workers {
			worker.mutex.RLock()
			if !worker.isRunning && worker.lastActive.Before(oldestTime) {
				oldestTime = worker.lastActive
				oldestWorker = worker
				oldestID = id
			}
			worker.mutex.RUnlock()
		}

		// 如果找到空闲超过idleTimeout的worker，则停止它
		if oldestWorker != nil && time.Since(oldestTime) > eq.idleTimeout {
			eq.stopWorker(oldestID)
			Logger.Info("自动缩容worker",
				"stopped_worker_id", oldestID,
				"total_workers", len(eq.workers),
				"queue_length", queueLength,
				"idle_time", time.Since(oldestTime))
		}
	}
}

// Stop 停止邮件队列
func (eq *EmailQueue) Stop() {
	if eq.cancel != nil {
		Logger.Info("正在停止邮件队列...")

		// 发送停止信号
		eq.cancel()

		// 设置超时等待，避免无限期阻塞
		done := make(chan struct{})
		go func() {
			eq.wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			Logger.Info("邮件队列已正常停止")
		case <-time.After(10 * time.Second):
			Logger.Warn("邮件队列停止超时，强制退出")
		}
	}
}

// GetQueueStats 获取队列统计信息
func (eq *EmailQueue) GetQueueStats() (map[string]interface{}, error) {
	if eq == nil || eq.redis == nil {
		return map[string]interface{}{
			"status":          "disabled",
			"worker_count":    0,
			"queue_size":      0,
			"failed_size":     0,
			"processed_total": 0,
			"failed_total":    0,
		}, nil
	}

	queueLen, err := eq.redis.client.LLen(eq.ctx, eq.queueKey).Result()
	if err != nil {
		return nil, err
	}

	failedLen, err := eq.redis.client.LLen(eq.ctx, eq.failKey).Result()
	if err != nil {
		return nil, err
	}

	eq.statsMutex.RLock()
	eq.workerMutex.RLock() // 读取动态worker信息需要锁定
	stats := map[string]interface{}{
		"status":          "active",
		"worker_count":    len(eq.workers),
		"min_workers":     eq.minWorkers,
		"max_workers":     eq.maxWorkers,
		"queue_size":      queueLen,
		"failed_size":     failedLen,
		"processed_total": eq.stats.ProcessedTotal,
		"failed_total":    eq.stats.FailedTotal,
		"queue_key":       eq.queueKey,
		"fail_key":        eq.failKey,
	}
	eq.workerMutex.RUnlock()
	eq.statsMutex.RUnlock()

	return stats, nil
}

// StopEmailQueue 停止邮件队列
func StopEmailQueue() {
	if EmailQueueInstance != nil {
		Logger.Info("正在停止邮件队列...")
		EmailQueueInstance.cancel()
		EmailQueueInstance.wg.Wait()
		Logger.Info("邮件队列已停止")
	}
}

// generateTaskID 生成任务ID
func generateTaskID() string {
	return fmt.Sprintf("email_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
}

// incrementProcessedCount 增加处理计数
func (eq *EmailQueue) incrementProcessedCount() {
	eq.statsMutex.Lock()
	eq.stats.ProcessedTotal++
	eq.statsMutex.Unlock()
}

// incrementFailedCount 增加失败计数
func (eq *EmailQueue) incrementFailedCount() {
	eq.statsMutex.Lock()
	eq.stats.FailedTotal++
	eq.statsMutex.Unlock()
}

// RetryFailedEmails 重试失败的邮件
func (eq *EmailQueue) RetryFailedEmails() (int, error) {
	if eq.redis == nil || !eq.redis.IsAvailable() {
		return 0, fmt.Errorf("Redis不可用")
	}

	ctx := context.Background()
	count := 0

	for {
		// 从失败队列获取任务
		result, err := eq.redis.client.BRPop(ctx, 1*time.Second, eq.failKey).Result()
		if err != nil {
			if err.Error() == "redis: nil" {
				break // 队列为空
			}
			return count, fmt.Errorf("获取失败任务失败: %v", err)
		}

		if len(result) < 2 {
			continue
		}

		// 解析失败任务
		var failedTask map[string]interface{}
		if err := json.Unmarshal([]byte(result[1]), &failedTask); err != nil {
			Logger.Error("解析失败任务失败", "err", err)
			continue
		}

		// 提取原始任务
		taskData, ok := failedTask["task"]
		if !ok {
			continue
		}

		taskJSON, err := json.Marshal(taskData)
		if err != nil {
			Logger.Error("序列化任务失败", "err", err)
			continue
		}

		var task EmailTask
		if err := json.Unmarshal(taskJSON, &task); err != nil {
			Logger.Error("反序列化任务失败", "err", err)
			continue
		}

		// 重置重试次数并重新入队
		task.Retry = 0
		if err := eq.Push(task); err != nil {
			Logger.Error("重新入队失败", "err", err)
			continue
		}

		count++
	}

	return count, nil
}

// ClearFailedEmails 清理失败队列
func (eq *EmailQueue) ClearFailedEmails() (int, error) {
	if eq.redis == nil || !eq.redis.IsAvailable() {
		return 0, fmt.Errorf("Redis不可用")
	}

	// 获取失败队列长度
	count, err := eq.redis.client.LLen(context.Background(), eq.failKey).Result()
	if err != nil {
		return 0, fmt.Errorf("获取失败队列长度失败: %v", err)
	}

	// 删除失败队列
	_, err = eq.redis.client.Del(context.Background(), eq.failKey).Result()
	if err != nil {
		return 0, fmt.Errorf("删除失败队列失败: %v", err)
	}

	return int(count), nil
}

// sendEmailSync 同步发送邮件（兜底方案）
func sendEmailSync(to, subject, body string) error {
	cfg := GetConfiguration()
	if !cfg.Smtp.Enabled {
		Logger.Debug("SMTP未启用，跳过邮件发送")
		return nil
	}

	// 暂时返回nil，避免循环导入
	// 实际发送逻辑将在controllers层调用
	Logger.Debug("同步发送邮件", "to", to, "subject", subject)
	return nil
}
