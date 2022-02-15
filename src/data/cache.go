package data

/*
dbsize
10
const INVITE_DATA_KEY = "string:invite_%d" 4045
active_device scard 200766
session_xxx 192328
xxx_session 192328
2
RcToken_ 190000
chat_no_reply_msg 0
chat_count_ 121000
chat_call_info_ 600
accost_count_ 68500
first_chat_time_ 50000
accost_tmp_ 0
chat_need_reply_ 79000
self_accost_record 0
chat_first_msg_time_ 5514
chat_valid_msg 9348
chat_join_hand 12000
match_ 0
video_free_count scard 117390
video_free_ 117000
video_free_set_ 102183
1
visit_homepage_count 72000
blind_click hash 600

const ONLINE_KEY = "online_set" hash 600
const LOCATION_KEY = "location_key" hash 14936
const USER_ACTIVE_KEY = "user_active" 27520 ttl
const MOST_GIFT_KEY = "most_gift_key" 0
const USER_ONEKEY_ACCOST_CD_KEY = "user_onekey_accost_cd" hash 4285

const JOIN_HANDS_INFO = "join_hands_info"  78000            //保存信息,且用来判断是否牵线过
const JOIN_HANDS_PAIR = "join_hands_pair"  6000  ttl         //保存信息,过期
const JOIN_HANDS_INFO_WOMAN = "join_hands_info_woman"  11900 //仅用来判断是否牵线过
const JOIN_HANDS_COUNT = "join_hands_count"   34000             //每日牵线次数
const JOIN_HANDS_COUNT_WOMAN = "join_hands_count_woman" 18209 //女每日牵线次数

1
const withdrawKey = `withdrawuid:%d:%s:%d` 300
const USER_SESSION_KEY = "user_session" hash 117900
const WE_CHAT_ASK_LIST = "wechat:ask_list"   hash 269 //存的内容是有哪些用户向自己索要了微信的数据 主要用作索要通知列表
const WE_CHAT_ASK_LIST2 = "wechat:ask_list2" hash 132 //存的内容是自己向哪些用户索要了微信的数据 主要用于消息超时处理
const WE_CHAT_ASK_SUM = "wechat:ask_sum"   hash  123//存的内容是索要通知总数量
const WE_CHAT_ASK_REC = "wechat:ask_rec"   hash 3088  //存的内容是索要通知新增数量
glamourRank zset 7465
wealthRank zset 10000
user:user_call_record_ list 1600 list
("week_%d%d%d", tm1.Year(), tm1.Month(), tm1.Day()) hash 6 hash 2000+

const GUEST_RECMD_HASH = "hash:guest_recmd" hash 38
const GUEST_INDEX_SET = "set:guest_index" set 1642
const GUEST_JOIN_HAND_INDEX = "set:guest_join_hand_index" set 2063
const GUEST_RECMD_HASH2 = "hash:guest_recmd2" hash 3
const GUEST_RECMD_HASH3 = "hash:guest_recmd3" hash 7
red_package_device set 146145
red_package_device_day7 set 18350

LIKE_MINE = "likemine_%d" // likemine_100012  52 set
MINE_LIKE = "minelike_%d" // minelike_100012  36 set
SEEN_MINE = "seenmine_%d" // seenmine_100012  160433 hash

LIKE_MINE_NEW = "likeminenew_%d" // likeminenew_100012 45069 hash
MINE_LIKE_NEW = "minelikenew_%d" // minelikenew_100012 20431 hash
SEEN_MINE_NEW = "seenminenew_%d" // seenminenew_100012 0

SEEN_MINE_REC     = "seenmine_rec"    // seenmine_rec hash 185689
LIKE_MINE_NEW_REC = "likeminenew_rec" // likeminenew_rec 0
MINE_LIKE_NEW_REC = "minelikenew_rec" // minelikenew_rec 0
fmt.Sprintf("%s_%d", keyWord, base.Sex)
const DYNAMIC_DAYS_CHARM = "dynamic_days_charm" hash 10329 // dynamic_days_charm 今日获得魅力值信息
fmt.Sprintf("CHAT_WORD_USE_COUNT_%d", req.Uid) 222 hash
1
redback

*/
/*
2021-10-21 00:35
# Memory
used_memory:10287453032
used_memory_human:9.58G
used_memory_rss:10602528768
used_memory_rss_human:9.87G
used_memory_peak:10294109880
used_memory_peak_human:9.59G
used_memory_peak_perc:99.94%
used_memory_overhead:169964666
used_memory_startup:42257448
used_memory_dataset:10117488366
used_memory_dataset_perc:98.77%

# Keyspace
db0:keys=3,expires=0,avg_ttl=0
db1:keys=513252,expires=48722,avg_ttl=318732433
db2:keys=794789,expires=630,avg_ttl=54640604
db3:keys=1,expires=0,avg_ttl=0
db4:keys=2451,expires=2448,avg_ttl=2651287
db8:keys=297,expires=297,avg_ttl=451069760
db10:keys=388844,expires=42102,avg_ttl=434380290

2021-10-22 00:50
# Memory
used_memory:10688989624
used_memory_human:9.95G
used_memory_rss:11020517376
used_memory_rss_human:10.26G
used_memory_peak:10691553128
used_memory_peak_human:9.96G
used_memory_peak_perc:99.98%
used_memory_overhead:169896118
used_memory_startup:42257448
used_memory_dataset:10519093506
used_memory_dataset_perc:98.81%

# Keyspace
db0:keys=3,expires=0,avg_ttl=0
db1:keys=513109,expires=43928,avg_ttl=343210785
db2:keys=806012,expires=561,avg_ttl=56997534
db3:keys=1,expires=0,avg_ttl=0
db4:keys=1591,expires=1588,avg_ttl=3041901
db8:keys=285,expires=285,avg_ttl=435120838
db10:keys=394072,expires=48056,avg_ttl=378981002

2021-10-24 00:50
# Memory
used_memory:11206819120
used_memory_human:10.44G
used_memory_rss:11544272896
used_memory_rss_human:10.75G
used_memory_peak:11210248208
used_memory_peak_human:10.44G
used_memory_peak_perc:99.97%
used_memory_overhead:170275960
used_memory_startup:42257448
used_memory_dataset:11036543160
used_memory_dataset_perc:98.87%


# Keyspace
db0:keys=3,expires=0,avg_ttl=0
db1:keys=513362,expires=39106,avg_ttl=312601357
db2:keys=819795,expires=402,avg_ttl=53449033
db3:keys=1,expires=0,avg_ttl=0
db4:keys=1177,expires=1174,avg_ttl=4319269
db8:keys=273,expires=273,avg_ttl=407053060
db10:keys=390215,expires=44953,avg_ttl=331004093


2021-10-25 00:40
# Memory
used_memory:11458846168
used_memory_human:10.67G
used_memory_rss:11813797888
used_memory_rss_human:11.00G
used_memory_peak:11460356632
used_memory_peak_human:10.67G
used_memory_peak_perc:99.99%
used_memory_overhead:170274002
used_memory_startup:42257448
used_memory_dataset:11288572166
used_memory_dataset_perc:98.89%

# Keyspace
db0:keys=3,expires=0,avg_ttl=0
db1:keys=513802,expires=37448,avg_ttl=324570095
db2:keys=826913,expires=358,avg_ttl=53104405
db3:keys=1,expires=0,avg_ttl=0
db4:keys=589,expires=586,avg_ttl=5387032
db8:keys=268,expires=268,avg_ttl=412593164
db10:keys=386795,expires=41895,avg_ttl=327466681
*/
