#!/bin/bash

source setup.sh

eventTypeId=$1
if [ "$eventTypeId" == "" ]
then
    echo "error: \$eventTypeId missing"
    exit 1
fi

serviceId=$2
if [ "$serviceId" == "" ]
then
    echo "error: \$serviceId missing"
    exit 1
fi

args='{\"name\":\"ME\", \"count\":5}'
#echo $args

json='{
    "title": "High Severity",
    "name": "High Severity",
    "enabled": true,
    "condition": {
        "eventTypeIds": ["'"$eventTypeId"'"],
        "query": {
            "query": {
                "match": {"data.beta": 17}
            }
        }
    },
    "job": {
        "createdBy": "sdgsreg",
        "jobType":{
            "type": "execute-service",
            "data": {
                "dataInputs": {
                 "": {
                    "content": "'"$args"'",
                    "type":     "body",
                    "mimeType": "application/json"
                }
                },
                "dataOutput": [
                    {
                        "mimeType": "application/json",
                      "type":     "text"
                    }
                ],
                "serviceId": "'"$serviceId"'"
            }
        }
    }
}'


# inside dataInputs{}

#echo POST /trigger
#echo "$json"

ret=$($curl -S -s -XPOST -d "$json" "$workflowurl"/trigger)

#echo RETURN:
#echo "$ret"

echo `extract triggerId "$ret"`
