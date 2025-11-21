// src/utils/chatWebSocket.ts

let ws: WebSocket | null = null;
let pingInterval: number | null = null;
let isConnecting = false;

export const WS_CHAT_MTH = [
  'MESSAGE_RECEIVED',
  'USER_TYPING',
  'TYPING_STOP',
  'USER_JOINED',
  'USER_LEFT',
  'MESSAGE_READ',
] as const;

export type WSChatMethod = (typeof WS_CHAT_MTH)[number];

const getWebSocketUrl = (chatId: string, token: string): string => {
  const INJECTED_API_BASE_URL = import.meta.env.VITE_API_BASE_URL;

  if (INJECTED_API_BASE_URL) {
    const isLocalDevelopment = INJECTED_API_BASE_URL.includes('localhost');

    if (isLocalDevelopment) {
      // Local: Chat Service 직접 연결
      return `ws://localhost:8001/api/chats/ws/${chatId}?token=${encodeURIComponent(token)}`;
    }

    // 운영: ALB를 통한 라우팅
    const protocol = INJECTED_API_BASE_URL.startsWith('https') ? 'wss:' : 'ws:';
    const host = INJECTED_API_BASE_URL.replace(/^https?:\/\//, '');

    // 🔥 /api/chats/ws/:chatId
    return `${protocol}//${host}/api/chats/ws/${chatId}?token=${encodeURIComponent(token)}`;
  }

  // Fallback
  const host = window.location.host;

  if (host.includes('localhost') || host.includes('127.0.0.1')) {
    return `ws://localhost:8001/api/chats/ws/${chatId}?token=${encodeURIComponent(token)}`;
  }

  return `wss://api.wealist.co.kr/api/chats/ws/${chatId}?token=${encodeURIComponent(token)}`;
};

export const connectChatWebSocket = (chatId: string, onMessage: (data: any) => void) => {
  // 🔥 이미 연결 중이면 무시
  if (isConnecting) {
    console.log('⚠️ [Chat WS] 이미 연결 중입니다.');
    return;
  }

  // 🔥 기존 연결 정리
  if (ws) {
    if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
      console.log('🔌 [Chat WS] 기존 연결 종료 중...');
      ws.close();
    }
    ws = null;
  }

  if (pingInterval) {
    clearInterval(pingInterval);
    pingInterval = null;
  }

  let reconnectAttempts = 0;
  const maxReconnectAttempts = 5;
  const reconnectDelay = 3000;

  const connect = () => {
    const token = localStorage.getItem('accessToken');
    if (!token) {
      console.error('❌ [Chat WS] No access token');
      isConnecting = false;
      return;
    }

    const wsUrl = getWebSocketUrl(chatId, token);
    console.log('🔌 [Chat WS] 연결 시도:', wsUrl);

    isConnecting = true;
    ws = new WebSocket(wsUrl);

    ws.onopen = () => {
      console.log('✅ [Chat WS] 연결 성공!');
      isConnecting = false;
      reconnectAttempts = 0;

      // 🔥 Ping 시작 (Redis 온라인 상태 유지)
      pingInterval = window.setInterval(() => {
        if (ws && ws.readyState === WebSocket.OPEN) {
          try {
            ws.send(JSON.stringify({ type: 'ping' }));
            console.log('🏓 [Chat WS] Ping 전송');
          } catch (error) {
            console.error('❌ [Chat WS] Ping 전송 실패:', error);
          }
        }
      }, 30000); // 30초마다
    };

    ws.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data);

        if (data.type === 'pong') {
          console.log('🏓 [Chat WS] Pong 수신');
          return;
        }

        console.log('📨 [Chat WS] 메시지 수신:', data);
        onMessage(data);
      } catch (error) {
        console.error('❌ [Chat WS] 메시지 파싱 실패:', error);
      }
    };

    ws.onerror = (e) => {
      console.error('❌ [Chat WS] 에러:', e);
      isConnecting = false;
    };

    ws.onclose = (event) => {
      console.log(`🔌 [Chat WS] 연결 닫힘: ${event.code} ${event.reason}`);
      isConnecting = false;

      // Ping 정리
      if (pingInterval) {
        clearInterval(pingInterval);
        pingInterval = null;
      }

      // 🔥 정상 종료(1000)가 아니면 재연결
      if (event.code !== 1000 && reconnectAttempts < maxReconnectAttempts) {
        reconnectAttempts++;
        console.log(`🔄 [Chat WS] 재연결 시도 ${reconnectAttempts}/${maxReconnectAttempts}...`);
        setTimeout(connect, reconnectDelay);
      } else if (reconnectAttempts >= maxReconnectAttempts) {
        console.error('❌ [Chat WS] 최대 재연결 시도 초과');
      }
    };
  };

  connect();
};

export const disconnectChatWebSocket = () => {
  console.log('🔌 [Chat WS] 연결 해제 시도');

  // 🔥 Ping 정리
  if (pingInterval) {
    clearInterval(pingInterval);
    pingInterval = null;
  }

  // 🔥 WebSocket 정리
  if (ws) {
    if (ws.readyState === WebSocket.OPEN) {
      ws.close(1000, 'User disconnected');
      console.log('✅ [Chat WS] 정상 종료');
    } else if (ws.readyState === WebSocket.CONNECTING) {
      ws.close();
      console.log('⚠️ [Chat WS] 연결 중 강제 종료');
    }
    ws = null;
  }

  isConnecting = false;
};

// 🔥 메시지 전송 헬퍼
export const sendChatMessage = (message: any) => {
  if (!ws || ws.readyState !== WebSocket.OPEN) {
    console.warn('⚠️ [Chat WS] WebSocket not connected');
    return false;
  }

  try {
    ws.send(JSON.stringify(message));
    console.log('📤 [Chat WS] 메시지 전송:', message);
    return true;
  } catch (error) {
    console.error('❌ [Chat WS] 전송 실패:', error);
    return false;
  }
};

// 🔥 WebSocket 연결 상태 확인
export const isChatWebSocketConnected = (): boolean => {
  return ws !== null && ws.readyState === WebSocket.OPEN;
};
