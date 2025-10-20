#!/bin/bash
# SSL证书管理工具脚本
# 支持生成私钥、自签名证书和Let's Encrypt证书申请

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_DOMAIN="localhost"
DEFAULT_EMAIL="admin@example.com"
DEFAULT_CERT_DIR="certs"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    echo "SSL证书管理工具"
    echo ""
    echo "用法:"
    echo "  $0 [命令] [选项]"
    echo ""
    echo "命令:"
    echo "  key       生成RSA私钥"
    echo "  self      生成自签名证书"
    echo "  lets      申请Let's Encrypt证书"
    echo "  config    生成配置示例"
    echo "  help      显示帮助信息"
    echo ""
    echo "选项:"
    echo "  -d, --domain    域名 (默认: $DEFAULT_DOMAIN)"
    echo "  -e, --email     邮箱地址 (默认: $DEFAULT_EMAIL)"
    echo "  -o, --output    输出目录 (默认: $DEFAULT_CERT_DIR)"
    echo "  -k, --keyfile   私钥文件名 (默认: server.key)"
    echo "  -c, --certfile  证书文件名 (默认: server.crt)"
    echo ""
    echo "示例:"
    echo "  $0 key -k mysite.key"
    echo "  $0 self -d example.com -k mysite.key -c mysite.crt"
    echo "  $0 lets -d example.com -e admin@example.com"
    echo ""
}

# 生成RSA私钥
generate_key() {
    local keyfile="$1"
    local keysize="${2:-2048}"
    
    if [[ -f "$keyfile" ]]; then
        print_warning "私钥文件 $keyfile 已存在"
        read -p "是否覆盖? (y/N): " confirm
        if [[ $confirm != [yY] ]]; then
            print_info "操作已取消"
            return 0
        fi
    fi
    
    print_info "生成 $keysize 位RSA私钥: $keyfile"
    
    if ! command -v openssl &> /dev/null; then
        print_error "未找到openssl命令，请先安装OpenSSL"
        return 1
    fi
    
    openssl genrsa -out "$keyfile" "$keysize"
    
    if [[ $? -eq 0 ]]; then
        chmod 600 "$keyfile"
        print_success "私钥生成成功: $keyfile"
        print_info "文件权限已设置为600"
        
        # 显示私钥信息
        echo ""
        echo "私钥信息:"
        openssl rsa -in "$keyfile" -text -noout | head -5
    else
        print_error "私钥生成失败"
        return 1
    fi
}

# 生成自签名证书
generate_self_signed() {
    local domain="$1"
    local keyfile="$2"
    local certfile="$3"
    local days="${4:-365}"
    
    print_info "生成自签名证书"
    print_info "域名: $domain"
    print_info "私钥: $keyfile"
    print_info "证书: $certfile"
    print_info "有效期: $days 天"
    
    # 检查私钥是否存在
    if [[ ! -f "$keyfile" ]]; then
        print_warning "私钥文件 $keyfile 不存在，将自动生成"
        generate_key "$keyfile"
    fi
    
    # 生成证书
    openssl req -new -x509 -key "$keyfile" -out "$certfile" -days "$days" \
        -subj "/C=CN/ST=State/L=City/O=Organization/OU=OrgUnit/CN=$domain"
    
    if [[ $? -eq 0 ]]; then
        chmod 644 "$certfile"
        print_success "自签名证书生成成功: $certfile"
        
        echo ""
        echo "证书信息:"
        openssl x509 -in "$certfile" -text -noout | grep -E "(Subject:|Not Before|Not After)"
        
        echo ""
        print_info "配置文件设置:"
        echo "[tls]"
        echo "enabled = true"
        echo "auto_cert = false"
        echo "cert_file = \"$certfile\""
        echo "key_file = \"$keyfile\""
        
        echo ""
        print_warning "注意: 自签名证书仅适用于开发测试环境"
    else
        print_error "证书生成失败"
        return 1
    fi
}

# 申请Let's Encrypt证书
apply_lets_encrypt() {
    local domain="$1"
    local email="$2"
    
    if [[ -z "$domain" || "$domain" == "localhost" ]]; then
        print_error "Let's Encrypt证书需要真实域名，不能使用localhost"
        return 1
    fi
    
    if [[ -z "$email" ]]; then
        print_error "Let's Encrypt证书需要提供邮箱地址"
        return 1
    fi
    
    print_info "申请Let's Encrypt证书"
    print_info "域名: $domain"
    print_info "邮箱: $email"
    
    # 检查certbot是否安装
    if ! command -v certbot &> /dev/null; then
        print_error "未找到certbot命令"
        echo ""
        echo "安装方法:"
        echo "Ubuntu/Debian: sudo apt install certbot"
        echo "CentOS/RHEL:   sudo yum install certbot"
        echo "macOS:         brew install certbot"
        return 1
    fi
    
    # 检查端口80是否可用
    if netstat -ln | grep -q ":80 "; then
        print_warning "端口80被占用，Let's Encrypt需要使用端口80进行验证"
        print_info "请确保端口80可用，或使用DNS验证方式"
    fi
    
    print_info "开始申请证书..."
    sudo certbot certonly --standalone -d "$domain" --email "$email" --agree-tos --non-interactive
    
    if [[ $? -eq 0 ]]; then
        local cert_path="/etc/letsencrypt/live/$domain"
        print_success "Let's Encrypt证书申请成功"
        
        echo ""
        print_info "证书文件位置:"
        echo "证书: $cert_path/fullchain.pem"
        echo "私钥: $cert_path/privkey.pem"
        
        echo ""
        print_info "配置文件设置:"
        echo "[tls]"
        echo "enabled = true"
        echo "auto_cert = true"
        echo "domain = \"$domain\""
        echo "email = \"$email\""
        
        echo ""
        print_info "或手动指定证书路径:"
        echo "[tls]"
        echo "enabled = true"
        echo "auto_cert = false"
        echo "cert_file = \"$cert_path/fullchain.pem\""
        echo "key_file = \"$cert_path/privkey.pem\""
        
        echo ""
        print_info "证书自动续期设置:"
        echo "添加到crontab (每月检查一次):"
        echo "0 2 1 * * /usr/bin/certbot renew --quiet"
    else
        print_error "证书申请失败"
        return 1
    fi
}

# 生成配置示例
generate_config() {
    cat << 'EOF'
# HTTPS配置示例

## 1. 使用自动Let's Encrypt证书 (推荐生产环境)
[tls]
enabled = true
auto_cert = true
domain = "yourdomain.com"
email = "admin@yourdomain.com"
cert_dir = "certs"

## 2. 使用手动证书
[tls]
enabled = true
auto_cert = false
cert_file = "server.crt"
key_file = "server.key"

## 3. 禁用HTTPS (仅HTTP)
[tls]
enabled = false

# 注意事项:
# - 自动证书需要域名解析到服务器IP
# - 自动证书需要开放80和443端口
# - 手动证书需要自行获取和更新
# - 开发测试可使用自签名证书
EOF
}

# 解析命令行参数
DOMAIN="$DEFAULT_DOMAIN"
EMAIL="$DEFAULT_EMAIL"
OUTPUT_DIR="$DEFAULT_CERT_DIR"
KEYFILE="server.key"
CERTFILE="server.crt"

while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--domain)
            DOMAIN="$2"
            shift 2
            ;;
        -e|--email)
            EMAIL="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -k|--keyfile)
            KEYFILE="$2"
            shift 2
            ;;
        -c|--certfile)
            CERTFILE="$2"
            shift 2
            ;;
        key|self|lets|config|help)
            COMMAND="$1"
            shift
            ;;
        *)
            echo "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 确保输出目录存在
if [[ ! -d "$OUTPUT_DIR" && "$COMMAND" != "help" && "$COMMAND" != "config" ]]; then
    mkdir -p "$OUTPUT_DIR"
    print_info "创建目录: $OUTPUT_DIR"
fi

# 切换到输出目录
if [[ "$COMMAND" != "help" && "$COMMAND" != "config" && "$COMMAND" != "lets" ]]; then
    cd "$OUTPUT_DIR"
fi

# 执行命令
case "$COMMAND" in
    key)
        generate_key "$KEYFILE"
        ;;
    self)
        generate_self_signed "$DOMAIN" "$KEYFILE" "$CERTFILE"
        ;;
    lets)
        apply_lets_encrypt "$DOMAIN" "$EMAIL"
        ;;
    config)
        generate_config
        ;;
    help|"")
        show_help
        ;;
    *)
        print_error "未知命令: $COMMAND"
        show_help
        exit 1
        ;;
esac