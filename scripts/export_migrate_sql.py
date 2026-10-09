# -*- coding: utf-8 -*-
"""从测试环境 MySQL 导出指定表数据，生成 migrate-data.sql（仅数据，不含表结构）。"""
import datetime
import decimal

import pymysql

TABLES = ['llm_config', 'alert_rules', 'business_groups', 'datasources',
          'notify_media', 'notify_rule', 'services']
OUT = 'sql/migrate-data.sql'
BATCH = 200

ESCAPE = {
    '\0': '\\0', '\\': '\\\\', "'": "\\'", '"': '\\"',
    '\n': '\\n', '\r': '\\r', '\x1a': '\\Z', '\b': '\\b', '\t': '\\t',
}


def esc(s):
    return "'" + ''.join(ESCAPE.get(c, c) for c in s) + "'"


def lit(v):
    if v is None:
        return 'NULL'
    if isinstance(v, bool):
        return '1' if v else '0'
    if isinstance(v, int):
        return str(v)
    if isinstance(v, (float, decimal.Decimal)):
        return str(v)
    if isinstance(v, (bytes, bytearray)):
        return '0x' + v.hex()
    if isinstance(v, datetime.datetime):
        s = v.strftime('%Y-%m-%d %H:%M:%S.%f') if v.microsecond else v.strftime('%Y-%m-%d %H:%M:%S')
        return esc(s)
    if isinstance(v, datetime.date):
        return esc(v.strftime('%Y-%m-%d'))
    if isinstance(v, datetime.time):
        return esc(str(v))
    if isinstance(v, str):
        return esc(v)
    return esc(str(v))


def main():
    conn = pymysql.connect(host='10.2.216.92', port=3306, user='root',
                           password='a.123456R+=', database='aiops',
                           charset='utf8mb4', connect_timeout=10)
    cur = conn.cursor()
    try:
        with open(OUT, 'w', encoding='utf-8', newline='\n') as f:
            f.write('-- AIOPS 业务配置数据迁移（从测试环境 10.2.216.92:3306/aiops 导出）\n')
            f.write('-- 仅数据，不含表结构；导入前目标库会执行 SET NAMES utf8mb4\n')
            f.write('-- 表: ' + ', '.join(TABLES) + '\n\n')
            f.write('SET NAMES utf8mb4;\nSET FOREIGN_KEY_CHECKS=0;\n\n')

            total = 0
            for t in TABLES:
                cur.execute('SHOW TABLES LIKE %s', (t,))
                if not cur.fetchone():
                    print('!! 源库缺少表 %s，已跳过' % t)
                    continue
                cur.execute('SELECT * FROM `%s`' % t)
                cols = [d[0] for d in cur.description]
                rows = cur.fetchall()
                collist = ', '.join('`%s`' % c for c in cols)
                f.write('-- ----------------------------\n')
                f.write('-- 表 `%s`  共 %d 行\n' % (t, len(rows)))
                f.write('-- ----------------------------\n')
                f.write('LOCK TABLES `%s` WRITE;\n' % t)
                f.write('/*!40000 ALTER TABLE `%s` DISABLE KEYS */;\n' % t)
                for i in range(0, len(rows), BATCH):
                    chunk = rows[i:i + BATCH]
                    values = ',\n'.join('(' + ', '.join(lit(v) for v in r) + ')' for r in chunk)
                    f.write('INSERT INTO `%s` (%s) VALUES\n%s;\n' % (t, collist, values))
                f.write('/*!40000 ALTER TABLE `%s` ENABLE KEYS */;\n' % t)
                f.write('UNLOCK TABLES;\n\n')
                print('%-18s %d 行' % (t, len(rows)))
                total += len(rows)
            f.write('SET FOREIGN_KEY_CHECKS=1;\n')
        print('\n完成: %s  共 %d 行数据' % (OUT, total))
    finally:
        conn.close()


if __name__ == '__main__':
    main()
