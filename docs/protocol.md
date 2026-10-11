# MaxIM Protocol Command Matrix

Status Legend:
- 🟢 Verified: Confirmed via network captures and binary analysis.
- 🟡 Identified: Found in binary / decompilation; parameters or behavior partially mapped or infered.
- 🔴 Unmapped: String identified. Syntax, direction, or role unconfirmed.

| Command           |  Dir  | Identified Arguments                                   | Expected Response(s)                         | Status | Spec Sheet |
|:------------------|:-----:|:-------------------------------------------------------|:---------------------------------------------|:------:|:-----------|
| `CLIENT_PROTOCOL` | C → S | `<version:int>`                                        | (None)                                       |   🟢   | TODO       |
| `CLIENT_TYPE`     | C → S | `<platform:str> ID=<id:hex>`                           | (None)                                       |   🟢   | TODO       |
| `LOGIN`           | C → S | `<user:str> <pass:str>`                                | `LOGIN_OK`, `LOGIN_FAILED`                   |   🟢   | TODO       |
| `REGISTER`        | C → S | (User details, see below)                              | `REGISTER_OK`, `REGISTER_FAILED`             |   🟢   | TODO       |
| `QUIT`            | C → S | (None)                                                 | (TCP Disconnect)                             |   🟢   | TODO       |
| `CHAT`            | C → S | (None)                                                 | `REPORT_USERS/CHANNELS`, `BUDDY/IGNORE_LIST` |   🟢   | TODO       |
| `REPORT_CHANNELS` | S → C | `<channels:space separated list>...`                   | (None)                                       |   🟢   | TODO       |
| `BUDDY_LIST`      | S → C | `<users:space separated list>...`                      | (None)                                       |   🟢   | TODO       |
| `IGNORE_LIST`     | S → C | `<users:space separated list>...`                      | (None)                                       |   🟢   | TODO       |
| `FORUMS`          | C → S | (None)                                                 | `GET_FORUM_LIST`                             |   🟢   | TODO       |
| `TOPICS`          | C → S | `<forum_id:int>`                                       | `GET_TOPIC_LIST`                             |   🟢   | TODO       |
| `POSTS`           | C → S | `<forum_id:int> <topic_id:int>`                        | `GET_POSTS_LIST`                             |   🟢   | TODO       |
| `JOIN`            | C → S | `<room:str>`                                           | `REPORT_USERS`                               |   🟢   | TODO       |
| `REPORT_USERS`    | S → C | `<channel:str> <users:space separated list>...`        | (None)                                       |   🟢   | TODO       |
| `SAY`             | C → S | `<message:str>`                                        | `USER_MSG`                                   |   🟢   | TODO       |
| `MSG`             | C → S | `<user:str> <msg:str>`                                 | `USER_MSG    `                               |   🟢   | TODO       |
| `CREATE`          | C → S | `<room:str>`                                           | `CHANNEL_ADDED`                              |   🟢   | TODO       |
| `WHOIS`           | C → S | `<user:str>`                                           | `SERVER_MSG`                                 |   🟢   | TODO       |
| `REPORT`          | C → S | `<user:str>`                                           | (None)                                       |   🟢   | TODO       |
| `IGNORE ADD`      | C → S | `<user:str>`                                           | `IGNORE_ADD`                                 |   🟢   | TODO       |
| `IGNORE REMOVE`   | C → S | `<user:str>`                                           | `IGNORE_DEL`                                 |   🟢   | TODO       |
| `BUDDY ADD`       | C → S | `<user:str>`                                           | `BUDDY_ADD`                                  |   🟢   | TODO       |
| `BUDDY REMOVE`    | C → S | `<user:str>`                                           | `BUDDY_DEL`                                  |   🟢   | TODO       |
| `LOGIN_OK`        | S → C | (None)                                                 | (None)                                       |   🟢   | TODO       |
| `LOGIN_FAIL`      | S → C | `<reason:str>`                                         | (None)                                       |   🟢   | TODO       |
| `REGISTER_OK`     | S → C | (None)                                                 | (None)                                       |   🟢   | TODO       |
| `REGISTER_FAIL`   | S → C | `<reason:str>`                                         | (None)                                       |   🟢   | TODO       |
| `CHANNEL_ADDED`   | S → C | `<room:str>`                                           | (None)                                       |   🟢   | TODO       |
| `CHANNEL_REMOVED` | S → C | `<room:str>`                                           | (None)                                       |   🟢   | TODO       |
| `IGNORE_ADD`      | S → C | `<user:str>`                                           | (None)                                       |   🟢   | TODO       |
| `IGNORE_DEL`      | S → C | `<user:str>`                                           | (None)                                       |   🟢   | TODO       |
| `BUDDY_ADD`       | S → C | `<user:str>`                                           | (None)                                       |   🟢   | TODO       |
| `BUDDY_DEL`       | S → C | `<user:str>`                                           | (None)                                       |   🟢   | TODO       |
| `BUDDY_STATUS`    | S → C | `<user:str> <status:str:+→`                            | (None)                                       |   🟢   | TODO       |
| `USER_MSG`        | S → C | `<msg:str>`                                            | (None)                                       |   🟢   | TODO       |
| `USER_JOIN`       | S → C | `<user:str>`                                           | (None)                                       |   🟢   | TODO       |
| `USER_LEAVE`      | S → C | `<user:str>`                                           | (None)                                       |   🟢   | TODO       |
| `SERVER_MSG`      | S → C | `<msg:str>`                                            | (None)                                       |   🟢   | TODO       |
| `SERVER_ERR`      | S → C | `<msg:str>`                                            | (None)                                       |   🟢   | TODO       |
| `DISPLAY_MSG`     | S → C | `<msg:str>`                                            | (None)                                       |   🟢   | TODO       |
| `NEWDETAILS`      | C → S | (User details, see below)                              | `USER_DETAILS` → `NEWDETAILS_OK`             |   🟢   | TODO       |
| `USER_DETAILS`    | S → C | (User details, see below)                              | (None)                                       |   🟢   | TODO       |
| `NEWDETAILS_OK`   | S → C | (None)                                                 | (None)                                       |   🟢   | TODO       |
| `PING`            | C ⇄ S | (None)                                                 | `PONG`                                       |   🟢   | TODO       |
| `PONG`            | C ⇄ S | (None)                                                 | (None)                                       |   🟢   | TODO       |
| `GET_FORUM_LIST`  | S → C | (space separated forum info, see below)                | (None)                                       |   🟡   | TODO       |
| `GET_TOPIC_LIST`  | S → C | (space separated topic info, see below)                | (None)                                       |   🟡   | TODO       |
| `GET_POSTS_LIST`  | S → C | `<pages:int> <(space separated post info, see below)>` | (None)                                       |   🟡   | TODO       |
| `GETAVATARS`      |   ?   | `?`                                                    | `?`                                          |   🔴   |            |
| `AVATARIMAGE`     |   ?   | `?`                                                    | `?`                                          |   🔴   |            |
| `SMILEPACKAGE`    |   ?   | `?`                                                    | `?`                                          |   🔴   |            |
| `ADDTOPIC`        |   ?   | `?`                                                    | `?`                                          |   🔴   |            |
| `ADDPOST`         |   ?   | `?`                                                    | `?`                                          |   🔴   |            |


User details → `<username:str> <password:str> <fullname:str> <gender:str> <location:str> <DOB:str:dd/mm/yy> <email:str:urlencoded> <profile:str> <sig:str> <keep_avatar/get_random:str:y/n>`

Forum info → `<id:int>+<name:str>+<unused???:str>+<last_post:str>+<topics:int>+<posts:int>+<type:str:N/L/O>`
-- N → New activity
-- O → Open?
-- L → Locked
-- If type is C, it is a "Category"
Topic info → `<id:int>+<name:str>+<unused???:str>+<author:str>+<posts:int>+<unused???:str>+<unused???:str>+<last_post:str>+<type:str:M/L/O>`
-- M → ??
-- O → Open
-- C → Closed
Post info → `<message:str>+<date:str:dd/mm/yy>+<author:str>+<unused???:str>+<unused???:str>+<user_post_count:int>+<unused???:str>+<user_id:int>`

All str fields can be urlencoded to encode special characters like spaces!
