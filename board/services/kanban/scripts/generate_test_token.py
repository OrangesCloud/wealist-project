#!/usr/bin/env python3
"""
테스트용 JWT 토큰 생성 스크립트

Member 서비스가 아직 없을 때 Kanban API를 테스트하기 위한 JWT 토큰을 생성합니다.

사용법:
    python scripts/generate_test_token.py
    python scripts/generate_test_token.py --user-id 2
    python scripts/generate_test_token.py --user-id 1 --expire-days 30
"""
import sys
import os
from datetime import datetime, timedelta
from jose import jwt
import argparse

# 프로젝트 루트를 PYTHONPATH에 추가
sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from app.config import settings


def generate_token(user_id: int, expire_days: int = 7) -> str:
    """
    테스트용 JWT 토큰 생성

    Args:
        user_id: 사용자 ID
        expire_days: 토큰 만료일 (기본 7일)

    Returns:
        str: JWT 토큰
    """
    # 토큰 페이로드
    payload = {
        "sub": str(user_id),  # user_id
        "exp": datetime.utcnow() + timedelta(days=expire_days),
        "iat": datetime.utcnow(),
        "type": "access_token"
    }

    # JWT 토큰 생성
    token = jwt.encode(
        payload,
        settings.SECRET_KEY,
        algorithm=settings.ALGORITHM
    )

    return token


def main():
    parser = argparse.ArgumentParser(
        description="테스트용 JWT 토큰 생성"
    )
    parser.add_argument(
        "--user-id",
        type=int,
        default=1,
        help="사용자 ID (기본값: 1)"
    )
    parser.add_argument(
        "--expire-days",
        type=int,
        default=7,
        help="토큰 만료일 (기본값: 7일)"
    )

    args = parser.parse_args()

    # 토큰 생성
    token = generate_token(args.user_id, args.expire_days)

    print("=" * 80)
    print("테스트용 JWT 토큰 생성 완료")
    print("=" * 80)
    print(f"\nUser ID: {args.user_id}")
    print(f"만료: {args.expire_days}일 후")
    print(f"\n토큰:\n{token}")
    print("\n" + "=" * 80)
    print("\n사용 방법:")
    print("\n1. cURL:")
    print(f'   curl -H "Authorization: Bearer {token}" \\')
    print('        http://localhost:8000/api/workspaces/')
    print("\n2. HTTPie:")
    print(f'   http GET http://localhost:8000/api/workspaces/ \\')
    print(f'        "Authorization: Bearer {token}"')
    print("\n3. Swagger UI:")
    print("   - http://localhost:8000/docs 접속")
    print("   - 우측 상단 'Authorize' 버튼 클릭")
    print(f"   - 토큰 입력: {token}")
    print("=" * 80)


if __name__ == "__main__":
    main()
