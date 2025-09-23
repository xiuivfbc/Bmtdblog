/* 后台管理现代化交互脚本 */

$(document).ready(function () {
    'use strict';

    // 侧边栏菜单激活状态处理
    function setActiveMenu() {
        var currentPath = window.location.pathname;
        $('.sidebar-menu li').removeClass('active');

        // 精确匹配当前页面的菜单项
        var menuMappings = {
            '/admin/index': '/admin/index',
            '/admin': '/admin/index',
            '/admin/post': '/admin/post',
            '/admin/page': '/admin/page',
            '/admin/user': '/admin/user',
            '/admin/subscriber': '/admin/subscriber',
            '/admin/link': '/admin/link'
        };

        // 首先尝试精确匹配
        var targetHref = menuMappings[currentPath];
        if (targetHref) {
            $('.sidebar-menu a[href="' + targetHref + '"]').parent('li').addClass('active');
            return;
        }

        // 如果没有精确匹配，尝试部分匹配
        $('.sidebar-menu a').each(function () {
            var href = $(this).attr('href');
            if (href && href !== '/admin/index' && currentPath.indexOf(href) === 0) {
                $(this).parent('li').addClass('active');
                return false; // 找到第一个匹配项后停止
            }
        });

        // 如果还是没有匹配到，且在admin路径下，则默认激活控制台
        if (currentPath.indexOf('/admin') === 0) {
            $('.sidebar-menu a[href="/admin/index"]').parent('li').addClass('active');
        }
    }

    // 信息卡片hover效果增强
    function enhanceInfoBoxes() {
        $('.info-box').hover(
            function () {
                $(this).find('.info-box-icon').css('transform', 'scale(1.1)');
            },
            function () {
                $(this).find('.info-box-icon').css('transform', 'scale(1)');
            }
        );
    }

    // 表格行hover效果
    function enhanceTableRows() {
        $('.table tbody tr').hover(
            function () {
                $(this).css('background-color', 'rgba(102, 126, 234, 0.05)');
            },
            function () {
                $(this).css('background-color', '');
            }
        );
    }

    // 按钮点击效果
    function enhanceButtons() {
        $('.btn').on('click', function () {
            var $btn = $(this);
            var originalTransform = $btn.css('transform');

            $btn.css('transform', 'scale(0.95)');

            setTimeout(function () {
                $btn.css('transform', originalTransform);
            }, 150);
        });
    }

    // 表单输入框聚焦效果
    function enhanceFormInputs() {
        $('.form-control').on('focus', function () {
            $(this).parent('.form-group').addClass('focused');
        }).on('blur', function () {
            $(this).parent('.form-group').removeClass('focused');
        });
    }

    // 添加加载动画
    function addLoadingAnimation() {
        // 为表单提交添加加载状态
        $('form').on('submit', function () {
            var $form = $(this);
            var $submitBtn = $form.find('button[type="submit"], input[type="submit"]');

            if ($submitBtn.length) {
                var originalText = $submitBtn.text() || $submitBtn.val();
                $submitBtn.prop('disabled', true);

                if ($submitBtn.is('button')) {
                    $submitBtn.html('<i class="fa fa-spinner fa-spin"></i> 处理中...');
                } else {
                    $submitBtn.val('处理中...');
                }

                // 如果5秒后仍未完成，恢复按钮状态
                setTimeout(function () {
                    $submitBtn.prop('disabled', false);
                    if ($submitBtn.is('button')) {
                        $submitBtn.text(originalText);
                    } else {
                        $submitBtn.val(originalText);
                    }
                }, 5000);
            }
        });
    }

    // 添加确认对话框
    function addConfirmDialogs() {
        $('[data-confirm]').on('click', function (e) {
            var message = $(this).data('confirm') || '确定要执行此操作吗？';
            if (!confirm(message)) {
                e.preventDefault();
                return false;
            }
        });
    }

    // 添加工具提示
    function addTooltips() {
        $('[data-toggle="tooltip"]').tooltip();

        // 为操作按钮添加工具提示
        $('.btn-sm').each(function () {
            var $btn = $(this);
            if (!$btn.attr('title') && !$btn.data('original-title')) {
                var text = $btn.text().trim();
                if (text) {
                    $btn.attr('title', text).tooltip();
                }
            }
        });
    }

    // 数字动画效果
    function animateNumbers() {
        $('.info-box-number').each(function () {
            var $this = $(this);
            var target = parseInt($this.text()) || 0;
            var current = 0;
            var increment = target / 30;

            var timer = setInterval(function () {
                current += increment;
                if (current >= target) {
                    current = target;
                    clearInterval(timer);
                }
                $this.text(Math.floor(current));
            }, 50);
        });
    }

    // 搜索功能增强
    function enhanceSearch() {
        var $searchInput = $('.navbar-form input[type="text"]');
        if ($searchInput.length) {
            $searchInput.on('input', function () {
                var value = $(this).val().toLowerCase();
                if (value.length >= 2) {
                    // 这里可以添加实时搜索功能
                    // 目前只是视觉反馈
                    $(this).css('border-color', '#667eea');
                } else {
                    $(this).css('border-color', '');
                }
            });
        }
    }

    // 初始化所有功能
    function init() {
        setActiveMenu();
        enhanceInfoBoxes();
        enhanceTableRows();
        enhanceButtons();
        enhanceFormInputs();
        addLoadingAnimation();
        addConfirmDialogs();
        addTooltips();

        // 延迟执行数字动画，确保页面已完全加载
        setTimeout(animateNumbers, 500);

        enhanceSearch();
    }

    // 页面加载完成后初始化
    init();

    // 为动态加载的内容重新初始化
    $(document).ajaxComplete(function () {
        setTimeout(function () {
            enhanceTableRows();
            enhanceButtons();
            addTooltips();
        }, 100);
    });

    // 响应式侧边栏切换
    $(window).on('resize', function () {
        if ($(window).width() <= 768) {
            $('body').addClass('sidebar-collapse');
        } else {
            $('body').removeClass('sidebar-collapse');
        }
    });

    // 滚动到顶部按钮
    var $scrollTop = $('<div id="scroll-top" style="position: fixed; bottom: 20px; right: 20px; z-index: 1000; display: none;"><button class="btn btn-primary btn-sm" style="border-radius: 50%; width: 40px; height: 40px; padding: 0;"><i class="fa fa-chevron-up"></i></button></div>');
    $('body').append($scrollTop);

    $(window).on('scroll', function () {
        if ($(this).scrollTop() > 300) {
            $scrollTop.fadeIn();
        } else {
            $scrollTop.fadeOut();
        }
    });

    $scrollTop.on('click', function () {
        $('html, body').animate({
            scrollTop: 0
        }, 600);
    });
});

// 全局工具函数
window.AdminUtils = {
    // 显示成功消息
    showSuccess: function (message) {
        this.showAlert(message, 'success');
    },

    // 显示错误消息
    showError: function (message) {
        this.showAlert(message, 'danger');
    },

    // 显示警告消息
    showWarning: function (message) {
        this.showAlert(message, 'warning');
    },

    // 显示信息消息
    showInfo: function (message) {
        this.showAlert(message, 'info');
    },

    // 通用消息显示
    showAlert: function (message, type) {
        type = type || 'info';
        var alertHtml = '<div class="alert alert-' + type + ' alert-dismissible" style="position: fixed; top: 20px; right: 20px; z-index: 9999; min-width: 300px;">' +
            '<button type="button" class="close" data-dismiss="alert" aria-hidden="true">&times;</button>' +
            message +
            '</div>';

        $('body').append(alertHtml);

        // 3秒后自动消失
        setTimeout(function () {
            $('.alert').last().fadeOut(function () {
                $(this).remove();
            });
        }, 3000);
    }
};