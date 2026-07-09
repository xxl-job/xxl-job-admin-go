import { defineConfig } from 'vite'
import path from 'path'
import vue from '@vitejs/plugin-vue'
import vueJsx from '@vitejs/plugin-vue-jsx'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'
// https://vitejs.dev/config/
// base 可通过 VITE_BASE 环境变量覆盖（云主机部署到根路径用 '/'，子路径用 '/xxx/'）；
// 默认开发为 ''，生产为 '/vue-admin/'（GitHub Pages 子路径）
const base = process.env.VITE_BASE ?? (process.env.NODE_ENV === 'production' ? '/vue-admin/' : '')

export default defineConfig({
  base,
  server: {
    host: true,
    watch: { usePolling: true },
    hmr: true,
    proxy: {
      // admin REST API（登录、任务/执行器/日志 CRUD）
      '/api': {
        target: 'http://localhost:18080',
        changeOrigin: true
      },
      // ConnectRPC（执行器 + 未来前端 connect-web 直连）
      '/xxl.v1': {
        target: 'http://localhost:9090',
        changeOrigin: true
      }
    }
  },
  plugins: [
    vue(),
    vueJsx(),
    AutoImport({
      //imports: ['vue'],
      //dts: 'src/auto-import.d.ts',
      resolvers: [ElementPlusResolver()]
    }),
    Components({
      //dts: 'src/commponents.d.ts',
      resolvers: [ElementPlusResolver()]
    }),

  ],
  css: {
    // css预处理器
    preprocessorOptions: {
      scss: {
        // 引入 mixin.scss 这样就可以在全局中使用 mixin.scss中预定义的变量了
        // 给导入的路径最后加上 ;
        //additionalData: '@import "@/assets/main.scss";'
      }
    }
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src')
    },
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json', '.vue']
  },
  build: {
    // 代码分割配置
    rollupOptions: {
      output: {
        manualChunks: {
          // 将 Element Plus 相关代码单独打包
          'element-plus': ['element-plus'],
          // 将 Vue 相关代码单独打包
          vue: ['vue', 'vue-router', 'pinia'],
          // 将其他第三方库单独打包
          vendor: ['axios', '@vueuse/core'],
          // ECharts 独立分包（仅 dashboard 懒加载，独立缓存）
          echarts: ['echarts']
        }
      }
    },
    // 压缩配置
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true,
        drop_debugger: true
      }
    },
    // 打包文件大小限制
    chunkSizeWarningLimit: 1000
  }
})
