import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
    plugins: [vue()],
    server: {
        host: '0.0.0.0',
        port: 5173,
        strictPort: true,
        proxy: {
            // 前端请求 /api/v1/... -> 转发到后端
            '/api': {
                target: 'http://192.168.1.10', // 建议走 nginx 80
                changeOrigin: true,
                secure: false, // https 自签名证书场景也能代理；http 场景无副作用
                ws: true,
                // 当前路径本来就是 /api 开头，这里保持不变（等价于不改）
                rewrite: (path) => path.replace(/^\/api/, '/api'),

                // 调试代理问题时很有用
                configure: (proxy) => {
                    proxy.on('proxyReq', (proxyReq, req) => {
                        console.log('[proxy:req]', req.method, req.url, '->', proxyReq.protocol + '//' + proxyReq.host + proxyReq.path)
                    })
                    proxy.on('proxyRes', (proxyRes, req) => {
                        console.log('[proxy:res]', req.method, req.url, '<-', proxyRes.statusCode)
                    })
                    proxy.on('error', (err, req) => {
                        console.error('[proxy:error]', req.method, req.url, err.message)
                    })
                }
            }
        }
    }
})
