"""
User Service 클라이언트
weAlist-User 서비스와의 통신을 담당합니다.
"""

import httpx
import logging
from typing import Optional, Dict, Any
from uuid import UUID

logger = logging.getLogger(__name__)


class UserServiceClient:
    """
    User Service와 통신하는 HTTP 클라이언트
    """

    def __init__(self, base_url: str = "http://user-service:8080"):
        """
        Args:
            base_url: User Service의 기본 URL
        """
        self.base_url = base_url.rstrip("/")
        self.timeout = httpx.Timeout(10.0)

    async def verify_user(self, user_id: UUID, token: str) -> Optional[Dict[str, Any]]:
        """
        사용자 정보를 가져와서 검증합니다.

        Args:
            user_id: 검증할 사용자 ID
            token: JWT 액세스 토큰

        Returns:
            사용자 정보 딕셔너리 또는 None (실패 시)
        """
        headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        }

        try:
            async with httpx.AsyncClient(timeout=self.timeout) as client:
                response = await client.get(
                    f"{self.base_url}/api/users/{user_id}",
                    headers=headers
                )

                if response.status_code == 200:
                    user_data = response.json()
                    logger.info(f"User {user_id} verified successfully")
                    return user_data
                elif response.status_code == 404:
                    logger.warning(f"User {user_id} not found")
                    return None
                elif response.status_code == 401:
                    logger.warning(f"Unauthorized access for user {user_id}")
                    return None
                else:
                    logger.error(f"Unexpected status code {response.status_code} from User Service")
                    return None

        except httpx.TimeoutException:
            logger.error(f"Timeout while verifying user {user_id}")
            return None
        except httpx.RequestError as e:
            logger.error(f"Request error while verifying user {user_id}: {str(e)}")
            return None
        except Exception as e:
            logger.error(f"Unexpected error while verifying user {user_id}: {str(e)}")
            return None

    async def get_user_info(self, user_id: UUID, token: str) -> Optional[Dict[str, Any]]:
        """
        사용자 상세 정보를 가져옵니다.

        Args:
            user_id: 사용자 ID
            token: JWT 액세스 토큰

        Returns:
            사용자 정보 딕셔너리 또는 None (실패 시)
        """
        return await self.verify_user(user_id, token)

    async def check_user_exists(self, user_id: UUID, token: str) -> bool:
        """
        사용자가 존재하는지 확인합니다.

        Args:
            user_id: 확인할 사용자 ID
            token: JWT 액세스 토큰

        Returns:
            사용자 존재 여부
        """
        user_info = await self.verify_user(user_id, token)
        return user_info is not None

    async def get_users_bulk(self, user_ids: list[UUID], token: str) -> Dict[UUID, Dict[str, Any]]:
        """
        여러 사용자의 정보를 한 번에 가져옵니다.

        Args:
            user_ids: 사용자 ID 리스트
            token: JWT 액세스 토큰

        Returns:
            사용자 ID를 키로 하는 사용자 정보 딕셔너리
        """
        results = {}

        # TODO: User Service에 bulk API가 있다면 그것을 사용하는 것이 더 효율적입니다
        # 현재는 각각 요청을 보내는 방식으로 구현
        for user_id in user_ids:
            user_info = await self.verify_user(user_id, token)
            if user_info:
                results[user_id] = user_info

        return results


# 싱글톤 인스턴스
user_service_client = UserServiceClient()
