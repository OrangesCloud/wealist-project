import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// 💡 [수정] 호스트 환경 변수를 사용하되, Docker 환경을 위해 '0.0.0.0'을 기본값으로 설정
const devHost = process.env.VITE_HOST || '0.0.0.0';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],

  // 💡 Docker 컨테이너 환경에서 외부 접근 및 포트 고정을 위한 설정
  server: {
    host: devHost, // 환경 변수가 없으면 '0.0.0.0' (외부 접속 허용) 사용
    port: 3000, // 포트 3000 고정 (docker-compose 매핑 포트와 일치)
    hmr: {
      // 핫 리로딩이 컨테이너 내부/외부에서 정상 작동하도록 웹소켓 포트 명시
      clientPort: 3000,
    },
    // CORS 설정을 추가하여 로컬 환경에서 백엔드 API 접근을 허용할 수 있습니다 (선택적)
    // proxy: {
    //   '/api': {
    //     target: 'http://nginx',
    //     changeOrigin: true,
    //     rewrite: (path) => path.replace(/^\/api/, '/api'),
    //   },
    // },
  },

  // 💡 원래 설정 유지: 모듈 해석 확장자를 명시적으로 정의 (TSX/TS 파일이 누락되지 않도록)
  resolve: {
    extensions: ['.mjs', '.js', '.ts', '.jsx', '.tsx', '.json'],
  },
});
