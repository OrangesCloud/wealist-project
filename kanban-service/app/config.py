from pydantic_settings import BaseSettings
from pydantic import field_validator, Field
from functools import lru_cache
from typing import Union

class Settings(BaseSettings):
    DATABASE_URL: str
    REDIS_URL: str
    SECRET_KEY: str = "change-this-in-production"
    ALGORITHM: str = "HS512"  # User Service와 동일한 알고리즘 사용
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 30
    ENV: str = "development"
    DEBUG: bool = True
    PROJECT_NAME: str = "Kanban Service"
    VERSION: str = "1.0.0"
    CORS_ORIGINS: Union[str, list[str]] = Field(
        default=["http://localhost:3000", "http://localhost:8000"]
    )

    @field_validator("CORS_ORIGINS", mode="before")
    @classmethod
    def parse_cors_origins(cls, v):
        if isinstance(v, str):
            # 콤마로 구분된 문자열을 리스트로 변환
            return [origin.strip() for origin in v.split(",")]
        return v

    class Config:
        env_file = ".env"
        case_sensitive = False
        extra = "ignore"  # .env의 추가 필드 무시

@lru_cache
def get_settings() -> Settings:
    return Settings()

settings = get_settings()
