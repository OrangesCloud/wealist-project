## 그라파나 사용법
localhost:9092 admin/admin 으로 접속   

1. Home 에서 Add your first data source 클릭
2. 프로메테우스 선택
3. URL에 host.docker.internal:8080(유저 서비스) 입력

4. 대시보드 설정 왼쪽 대시보드에서 New 버튼에서 import 클릭 후 Upload Json file에 18812_rev4.json 넣기
5. Prometheus에서  3에서 설정한 프로메테우스 클릭후 완료

6. test-api 실행시키면 cpu 증가하는것 확인 가능  