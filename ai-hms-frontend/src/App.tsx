import { RouterProvider } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { AuthProvider } from '@/contexts/AuthContext'
import router from './router'
import './index.css'

export default function App() {
  return (
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: '#1677ff',
          colorError:   '#ff4d4f',
          colorWarning: '#faad14',
          colorSuccess: '#52c41a',
          colorInfo:    '#1677ff',
          borderRadius: 8,        // --radius-md
          borderRadiusLG: 12,     // --radius-lg
          borderRadiusSM: 4,      // --radius-sm
          fontSize: 14,           // --text-body
          fontFamily: '-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Hiragino Sans GB","Microsoft YaHei",sans-serif',
        },
        components: {
          Button: { borderRadius: 8 },
          Modal:  { borderRadiusLG: 12 },
          Table:  { borderRadius: 8, headerBg: '#f5f7fa' },
          Tag:    { borderRadiusSM: 4 },
        },
      }}
    >
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </ConfigProvider>
  )
}
