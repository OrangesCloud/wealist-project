// utils/websocket.ts
let ws: WebSocket | null = null;

// 💡 [추가] WebSocket 이벤트 타입 상수
export const WS_BOARD_MTH = [
  'BOARD_CREATED',
  'BOARD_UPDATED',
  'BOARD_MOVED',
  'BOARD_DELETED', // 💡 나중을 위해 미리 추가
] as const;

export type WSBoardMethod = (typeof WS_BOARD_MTH)[number];

export const connectWebSocket = (projectId: string, onMessage: (data: any) => void) => {
  if (ws) ws.close();

  const token = localStorage.getItem('accessToken');
  if (!token) {
    console.error('No access token');
    return;
  }

  const encodedToken = encodeURIComponent(token);

  ws = new WebSocket(`ws://localhost:8000/api/ws/project/${projectId}?token=${encodedToken}`);

  ws.onopen = () => console.log('✅ WebSocket 연결 성공!');
  ws.onmessage = (e) => {
    try {
      const data = JSON.parse(e.data);
      console.log('📨 [WS] 메시지 수신:', data);
      onMessage(data);
    } catch (error) {
      console.error('❌ [WS] 메시지 파싱 실패:', error);
    }
  };
  ws.onerror = (e) => console.error('❌ [WS] 에러:', e);
  ws.onclose = () => console.log('🔌 [WS] 연결 닫힘');
};

export const disconnectWebSocket = () => {
  if (ws) {
    ws.close();
    ws = null;
    console.log('🔌 [WS] 수동으로 연결 해제');
  }
};
