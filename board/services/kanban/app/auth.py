"""
인증 및 JWT 검증 모듈

Member 서비스와 동일한 SECRET_KEY를 공유하여 JWT 토큰 검증
User 서비스와 통신하여 사용자 존재 여부 확인
"""
from typing import Optional, Dict, Any
from fastapi import Depends, HTTPException, status
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from jose import JWTError, jwt
from app.config import settings
from app.logging_config import get_logger
from app.clients.user_client import user_service_client
from uuid import UUID

logger = get_logger(__name__)

# HTTP Bearer 토큰 스키마
security = HTTPBearer()


def get_current_user_id(
    credentials: HTTPAuthorizationCredentials = Depends(security)
) -> UUID:
    """
    JWT 토큰에서 user_id를 추출

    Member 서비스에서 발급한 JWT 토큰을 검증하고 user_id를 반환합니다.

    Args:
        credentials: HTTP Authorization Bearer 토큰

    Returns:
        UUID: 사용자 ID

    Raises:
        HTTPException: 토큰이 유효하지 않을 경우 401 에러
    """
    token = credentials.credentials

    try:
        # JWT 디코딩 (Member 서비스와 같은 SECRET_KEY 사용)
        payload = jwt.decode(
            token,
            settings.SECRET_KEY,
            algorithms=[settings.ALGORITHM]
        )

        # user_id 추출 (JWT의 "sub" 클레임)
        user_id: str = payload.get("sub")
        if user_id is None:
            logger.warning("Token missing 'sub' claim")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Could not validate credentials",
                headers={"WWW-Authenticate": "Bearer"},
            )

        logger.debug(f"JWT validated successfully for user_id: {user_id}")
        return UUID(user_id)

    except JWTError as e:
        logger.error(f"JWT validation failed: {str(e)}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not validate credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )


def get_current_user_id_optional(
    credentials: Optional[HTTPAuthorizationCredentials] = Depends(HTTPBearer(auto_error=False))
) -> Optional[UUID]:
    """
    선택적 인증 - 토큰이 있으면 user_id 반환, 없으면 None

    공개 API에서 사용 (인증된 사용자는 추가 정보 제공 가능)

    Args:
        credentials: HTTP Authorization Bearer 토큰 (선택)

    Returns:
        Optional[UUID]: 사용자 ID 또는 None
    """
    if credentials is None:
        return None

    try:
        payload = jwt.decode(
            credentials.credentials,
            settings.SECRET_KEY,
            algorithms=[settings.ALGORITHM]
        )
        user_id = payload.get("sub")
        return UUID(user_id) if user_id else None
    except JWTError:
        return None


async def get_verified_user_id(
    credentials: HTTPAuthorizationCredentials = Depends(security)
) -> UUID:
    """
    JWT 토큰 검증 + User Service에서 사용자 존재 확인

    Args:
        credentials: HTTP Authorization Bearer 토큰

    Returns:
        UUID: 검증된 사용자 ID

    Raises:
        HTTPException: 토큰이 유효하지 않거나 사용자가 존재하지 않을 경우 401 에러
    """
    # 1. JWT 토큰 검증
    user_id = get_current_user_id(credentials)
    token = credentials.credentials

    # 2. User Service에서 사용자 존재 확인
    try:
        user_exists = await user_service_client.check_user_exists(user_id, token)

        if not user_exists:
            logger.warning(f"User {user_id} not found in User Service")
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="User not found",
                headers={"WWW-Authenticate": "Bearer"},
            )

        logger.debug(f"User {user_id} verified with User Service")
        return user_id

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error verifying user with User Service: {str(e)}")
        # User Service와 통신 실패 시 JWT 검증만으로 허용 (fallback)
        logger.warning(f"Falling back to JWT-only verification for user {user_id}")
        return user_id


async def get_user_info(
    credentials: HTTPAuthorizationCredentials = Depends(security)
) -> Dict[str, Any]:
    """
    사용자 정보를 User Service에서 가져옵니다.

    Args:
        credentials: HTTP Authorization Bearer 토큰

    Returns:
        Dict[str, Any]: 사용자 정보

    Raises:
        HTTPException: 사용자 정보를 가져올 수 없을 경우
    """
    user_id = get_current_user_id(credentials)
    token = credentials.credentials

    user_info = await user_service_client.get_user_info(user_id, token)

    if not user_info:
        logger.warning(f"Could not fetch user info for {user_id}")
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Could not fetch user information",
            headers={"WWW-Authenticate": "Bearer"},
        )

    return user_info
