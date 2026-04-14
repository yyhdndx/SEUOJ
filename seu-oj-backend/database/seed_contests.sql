START TRANSACTION;

INSERT INTO contests (id,title,description,rule_type,start_time,end_time,is_public,allow_practice,ranklist_freeze_at,created_by,created_at,updated_at) VALUES
(1,'SEU Spring Warmup 2026','一场用于演示榜单、封榜与正式赛提交的进行中比赛。','acm','2026-04-07 18:00:00.000','2026-04-07 22:00:00.000',1,0,'2026-04-07 21:00:00.000',4,NOW(3),NOW(3)),
(2,'SEU April Qualifier Preview','一场即将开始的公开比赛，用于测试报名和列表展示。','acm','2026-04-08 19:00:00.000','2026-04-08 22:00:00.000',1,0,NULL,4,NOW(3),NOW(3)),
(3,'SEU March Mock Replay','一场已结束并开放补题的历史比赛。','acm','2026-04-05 19:00:00.000','2026-04-05 21:00:00.000',1,1,'2026-04-05 20:30:00.000',4,NOW(3),NOW(3));

INSERT INTO contest_problems (contest_id,problem_id,problem_code,display_order,created_at) VALUES
(1,3,'A',1,NOW(3)),
(2,3,'A',1,NOW(3)),
(3,3,'A',1,NOW(3));

INSERT INTO contest_registrations (contest_id,user_id,created_at) VALUES
(1,1,'2026-04-07 18:01:00.000'),
(1,2,'2026-04-07 18:02:00.000'),
(1,3,'2026-04-07 18:03:00.000'),
(2,1,NOW(3)),
(3,1,'2026-04-05 18:50:00.000'),
(3,2,'2026-04-05 18:52:00.000');

INSERT INTO contest_announcements (contest_id,title,content,is_pinned,created_by,created_at,updated_at) VALUES
(1,'比赛开始','Warmup 已开始，A 题为两数求和，样例与正式数据一致。',1,4,'2026-04-07 18:00:30.000','2026-04-07 18:00:30.000'),
(1,'封榜说明','本场比赛 21:00 封榜，封榜后的提交不会出现在公开榜单中，但管理员榜单仍可见。',0,4,'2026-04-07 20:40:00.000','2026-04-07 20:40:00.000'),
(3,'补题开放','Mock Replay 已开放 practice 模式，赛后提交不会影响正式榜单。',1,4,'2026-04-05 21:05:00.000','2026-04-05 21:05:00.000');

INSERT INTO submissions (user_id,problem_id,contest_id,is_practice,language,code,status,passed_count,total_count,runtime_ms,memory_kb,compile_info,error_message,created_at,judged_at) VALUES
(1,3,1,0,'cpp','#include <iostream>\nusing namespace std;\nint main(){int a,b;cin>>a>>b;cout<<a-b<<"\\n";return 0;}','Wrong Answer',0,1,2,256,NULL,'output mismatch','2026-04-07 18:20:00','2026-04-07 18:20:02'),
(1,3,1,0,'cpp','#include <iostream>\nusing namespace std;\nint main(){int a,b;cin>>a>>b;cout<<a+b<<"\\n";return 0;}','Accepted',1,1,1,256,NULL,NULL,'2026-04-07 18:32:00','2026-04-07 18:32:01'),
(2,3,1,0,'python3','a,b=map(int,input().split())\nprint(a+b)','Accepted',1,1,15,10240,NULL,NULL,'2026-04-07 18:40:00','2026-04-07 18:40:01'),
(3,3,1,0,'go','package main\nimport "fmt"\nfunc main(){var a,b int; fmt.Scan(&a,&b); fmt.Println(a-b)}','Wrong Answer',0,1,4,512,NULL,'output mismatch','2026-04-07 21:10:00','2026-04-07 21:10:02'),
(3,3,1,0,'go','package main\nimport "fmt"\nfunc main(){var a,b int; fmt.Scan(&a,&b); fmt.Println(a+b)}','Accepted',1,1,4,512,NULL,NULL,'2026-04-07 21:18:00','2026-04-07 21:18:01'),
(1,3,3,0,'cpp','#include <iostream>\nusing namespace std;\nint main(){int a,b;cin>>a>>b;cout<<a+b<<"\\n";return 0;}','Accepted',1,1,1,256,NULL,NULL,'2026-04-05 19:25:00','2026-04-05 19:25:01'),
(2,3,3,0,'java','import java.util.*; public class Main { public static void main(String[] args){ Scanner in=new Scanner(System.in); int a=in.nextInt(), b=in.nextInt(); System.out.println(a-b); }}','Wrong Answer',0,1,120,32768,NULL,'output mismatch','2026-04-05 19:40:00','2026-04-05 19:40:03'),
(2,3,3,1,'java','import java.util.*; public class Main { public static void main(String[] args){ Scanner in=new Scanner(System.in); int a=in.nextInt(), b=in.nextInt(); System.out.println(a+b); }}','Accepted',1,1,118,32768,NULL,NULL,'2026-04-06 10:00:00','2026-04-06 10:00:03');

COMMIT;
