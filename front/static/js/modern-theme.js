/**
 * 现代化主题交互增强脚本
 */

(function ($) {
    'use strict';

    // 页面加载完成后初始化
    $(document).ready(function () {
        initModernTheme();
    });

    function initModernTheme() {
        // 1. 页面加载动画
        initLoadingAnimations();

        // 2. 导航栏效果
        initNavbarEffects();

        // 3. 滚动效果
        initScrollEffects();

        // 4. 卡片悬停效果
        initCardHoverEffects();

        // 5. 平滑滚动
        initSmoothScrolling();

        // 6. 表单增强
        initFormEnhancements();

        // 7. 图片懒加载
        initImageLazyLoading();

        // 8. 主题切换（可选）
        initThemeToggle();
    }

    // 页面加载动画
    function initLoadingAnimations() {
        // 为文章卡片添加渐入动画
        $('.post-card, .sidebar-widget').each(function (index) {
            $(this).css({
                'opacity': '0',
                'transform': 'translateY(30px)'
            });

            setTimeout(() => {
                $(this).css({
                    'opacity': '1',
                    'transform': 'translateY(0)',
                    'transition': 'all 0.6s cubic-bezier(0.4, 0, 0.2, 1)'
                });
            }, index * 100);
        });
    }

    // 导航栏效果
    function initNavbarEffects() {
        const navbar = $('.navbar');
        let lastScrollTop = 0;

        $(window).scroll(function () {
            const scrollTop = $(this).scrollTop();

            // 滚动时改变导航栏样式
            if (scrollTop > 50) {
                navbar.addClass('navbar-scrolled');
            } else {
                navbar.removeClass('navbar-scrolled');
            }

            // 向下滚动时隐藏导航栏，向上滚动时显示
            if (scrollTop > lastScrollTop && scrollTop > 200) {
                navbar.css('transform', 'translateY(-100%)');
            } else {
                navbar.css('transform', 'translateY(0)');
            }

            lastScrollTop = scrollTop;
        });
    }

    // 滚动效果
    function initScrollEffects() {
        // 回到顶部按钮
        const backToTop = $('#back-to-top');

        if (backToTop.length === 0) {
            // 如果不存在，创建回到顶部按钮
            $('body').append(`
                <a href="#" id="back-to-top" class="back-to-top" style="display: none;">
                    <i class="fas fa-chevron-up"></i>
                </a>
            `);
        }

        $(window).scroll(function () {
            if ($(this).scrollTop() > 300) {
                $('#back-to-top').fadeIn();
            } else {
                $('#back-to-top').fadeOut();
            }
        });

        $('#back-to-top').click(function (e) {
            e.preventDefault();
            $('html, body').animate({ scrollTop: 0 }, 800);
        });

        // 滚动视差效果
        initParallaxEffect();
    }

    // 视差滚动效果
    function initParallaxEffect() {
        $(window).scroll(function () {
            const scrolled = $(this).scrollTop();
            const parallax = $('.container');
            const speed = scrolled * 0.1;

            // 为主容器添加轻微的视差效果
            if (parallax.length) {
                parallax.css('transform', `translateY(${speed}px)`);
            }
        });
    }

    // 卡片悬停效果
    function initCardHoverEffects() {
        // 文章卡片悬停效果
        $('.post-card').hover(
            function () {
                $(this).find('.articleInfo').css({
                    'transform': 'translateY(-5px) scale(1.02)',
                    'box-shadow': 'var(--shadow-lg)'
                });
            },
            function () {
                $(this).find('.articleInfo').css({
                    'transform': 'translateY(0) scale(1)',
                    'box-shadow': 'var(--shadow-md)'
                });
            }
        );

        // 侧边栏组件悬停效果
        $('.sidebar-widget').hover(
            function () {
                $(this).css({
                    'transform': 'translateY(-3px)',
                    'box-shadow': 'var(--shadow-lg)'
                });
            },
            function () {
                $(this).css({
                    'transform': 'translateY(0)',
                    'box-shadow': 'var(--shadow-sm)'
                });
            }
        );

        // 标签悬停效果
        $('.changeTag, .tag-cloud-item').hover(
            function () {
                $(this).css('transform', 'translateY(-2px) scale(1.05)');
            },
            function () {
                $(this).css('transform', 'translateY(0) scale(1)');
            }
        );
    }

    // 平滑滚动
    function initSmoothScrolling() {
        $('a[href^="#"]').on('click', function (event) {
            const target = $(this.getAttribute('href'));
            if (target.length) {
                event.preventDefault();
                $('html, body').stop().animate({
                    scrollTop: target.offset().top - 100
                }, 800, 'easeInOutQuad');
            }
        });
    }

    // 表单增强
    function initFormEnhancements() {
        // 输入框聚焦效果
        $('.form-control').on('focus', function () {
            $(this).parent().addClass('focused');
        }).on('blur', function () {
            if (!$(this).val()) {
                $(this).parent().removeClass('focused');
            }
        });

        // 为表单添加加载状态
        $('form').on('submit', function () {
            const submitBtn = $(this).find('button[type="submit"], input[type="submit"]');
            const originalText = submitBtn.text() || submitBtn.val();

            submitBtn.prop('disabled', true);
            if (submitBtn.is('button')) {
                submitBtn.html('<i class="fas fa-spinner fa-spin"></i> 提交中...');
            } else {
                submitBtn.val('提交中...');
            }

            // 如果表单验证失败，恢复按钮状态
            setTimeout(() => {
                submitBtn.prop('disabled', false);
                if (submitBtn.is('button')) {
                    submitBtn.html(originalText);
                } else {
                    submitBtn.val(originalText);
                }
            }, 3000);
        });
    }

    // 图片懒加载
    function initImageLazyLoading() {
        if ('IntersectionObserver' in window) {
            const imageObserver = new IntersectionObserver((entries, observer) => {
                entries.forEach(entry => {
                    if (entry.isIntersecting) {
                        const img = entry.target;
                        img.src = img.dataset.src;
                        img.classList.remove('lazy');
                        imageObserver.unobserve(img);
                    }
                });
            });

            document.querySelectorAll('img[data-src]').forEach(img => {
                imageObserver.observe(img);
            });
        }
    }

    // 主题切换功能
    function initThemeToggle() {
        // 检测系统主题偏好
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');

        // 可以在这里添加暗色主题切换功能
        // 暂时保留接口，以后扩展使用
    }

    // 实用工具函数
    const utils = {
        // 防抖函数
        debounce: function (func, wait, immediate) {
            let timeout;
            return function executedFunction() {
                const context = this;
                const args = arguments;
                const later = function () {
                    timeout = null;
                    if (!immediate) func.apply(context, args);
                };
                const callNow = immediate && !timeout;
                clearTimeout(timeout);
                timeout = setTimeout(later, wait);
                if (callNow) func.apply(context, args);
            };
        },

        // 节流函数
        throttle: function (func, limit) {
            let inThrottle;
            return function () {
                const args = arguments;
                const context = this;
                if (!inThrottle) {
                    func.apply(context, args);
                    inThrottle = true;
                    setTimeout(() => inThrottle = false, limit);
                }
            };
        }
    };

    // 将utils暴露到全局
    window.modernThemeUtils = utils;

})(jQuery);

// 添加自定义缓动函数
jQuery.easing.easeInOutQuad = function (x, t, b, c, d) {
    if ((t /= d / 2) < 1) return c / 2 * t * t + b;
    return -c / 2 * ((--t) * (t - 2) - 1) + b;
};

// 性能优化：使用 requestAnimationFrame
function smoothScrollTo(target, duration = 800) {
    const targetElement = typeof target === 'string' ? document.querySelector(target) : target;
    if (!targetElement) return;

    const targetPosition = targetElement.offsetTop - 100;
    const startPosition = window.pageYOffset;
    const distance = targetPosition - startPosition;
    let startTime = null;

    function animation(currentTime) {
        if (startTime === null) startTime = currentTime;
        const timeElapsed = currentTime - startTime;
        const run = ease(timeElapsed, startPosition, distance, duration);
        window.scrollTo(0, run);
        if (timeElapsed < duration) requestAnimationFrame(animation);
    }

    function ease(t, b, c, d) {
        t /= d / 2;
        if (t < 1) return c / 2 * t * t + b;
        t--;
        return -c / 2 * (t * (t - 2) - 1) + b;
    }

    requestAnimationFrame(animation);
}