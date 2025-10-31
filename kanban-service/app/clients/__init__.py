"""
클라이언트 모듈
외부 서비스와의 통신을 담당하는 클라이언트들을 정의합니다.
"""

from app.clients.user_client import UserServiceClient

__all__ = ["UserServiceClient"]
