#!/usr/bin/bash

cat << EOF
{
  "http": [
    {
      "listen": "127.0.0.1:1080",
      "forward": {
        "protocol": "https",
        "server": "<remote-server>:<remote-port>",
        "secret": "<username>:<password>",
        "match": [
          "a.b.c",
EOF

## wget https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt
## cat gfwlist.txt | base64 -d ｜ grep '^||'  > block.list

for block in $(cat block.list | sed 's/||//'); do
  echo "          \"$block\","
done

cat << EOF
          "x.y.z"
        ]
      }
    }
  ]
}
EOF