#!/usr/bin/env python3
"""
테스트용 JWT 토큰 생성 스크립트

프론트엔드 개발 시 API 테스트를 위한 JWT 토큰을 생성합니다.
"""
from jose import jwt
from datetime import datetime, timedelta
import argparse

# .env 파일의 설정과 동일
SECRET_KEY = "dev-secret-key-change-in-production"
ALGORITHM = "HS256"


def generate_token(user_id: int, expires_hours: int = 168) -> str:
    """
    JWT 토큰 생성

    Args:
        user_id: 사용자 ID
        expires_hours: 토큰 만료 시간 (시간 단위, 기본값: 168시간 = 7일)

    Returns:
        str: JWT 토큰
    """
    payload = {
        "sub": str(user_id),
        "exp": datetime.utcnow() + timedelta(hours=expires_hours),
        "iat": datetime.utcnow()
    }

    token = jwt.encode(payload, SECRET_KEY, algorithm=ALGORITHM)
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
        "--expires",
        type=int,
        default=168,
        help="토큰 만료 시간 (시간 단위, 기본값: 168 = 7일)"
    )
    parser.add_argument(
        "--bearer",
        action="store_true",
        help="Bearer 형식으로 출력"
    )

    args = parser.parse_args()

    token = generate_token(args.user_id, args.expires)

    print("\n" + "="*70)
    print(f"테스트용 JWT 토큰 생성 완료 (User ID: {args.user_id})")
    print(f"만료 시간: {args.expires}시간")
    print("="*70)
    print("\n토큰:")
    print("-"*70)

    if args.bearer:
        print(f"Bearer {token}")
    else:
        print(token)

    print("-"*70)
    print("\nAPI 호출 예시:")
    print("-"*70)
    print(f'curl -H "Authorization: Bearer {token}" \\')
    print('  http://localhost:8000/api/workspaces/')
    print("-"*70)
    print("\n프론트엔드에서 사용:")
    print("-"*70)
    print("headers: {")
    print(f'  "Authorization": "Bearer {token}"')
    print("}")
    print("-"*70)
    print()


if __name__ == "__main__":
    main()
