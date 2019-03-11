import json, requests, random

basetext = '{"appName":"APP_CHANGE","startTime":"2018-11-03T16:50:04-07:00","endTime":"2018-11-03T16:33:04-07:00", "metrics":{"current":{"error4xx":{"dataSourceType":"wavefront","parameters":{"end":ENDINGTIME,"endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/","query":"sum(ts(telegraf.http.server.requests.count, env=\\\"prd\\\" and app=\\\"REPLACE\\\" and status=\\\"5*\\\"))","start":STARTINGTIME,"step":60}},"latency":{"dataSourceType":"wavefront","parameters":{"end":ENDINGTIME,"endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/","query":"avg(ts(telegraf.http.server.requests.mean, env=\\\"prd\\\" and status=\\\"2*\\\" and app=\\\"REPLACE\\\"), app)","start":STARTINGTIME,"step":60}}},"historical":{"error4xx":{"dataSourceType":"wavefront","parameters":{"end":ENDINGTIME,"endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/","query":"sum(ts(telegraf.http.server.requests.count, env=\\\"prd\\\" and app=\\\"REPLACE\\\" and status=\\\"5*\\\"))","start":STARTINGTIME,"step":60}},"latency":{"dataSourceType":"wavefront","parameters":{"end":ENDINGTIME,"endpoint":"http://ab683be21d97f11e88e87023426427de-657499332.us-west-2.elb.amazonaws.com:9090/api/v1/","query":"avg(ts(telegraf.http.server.requests.mean, env=\\\"prd\\\" and status=\\\"2*\\\" and app=\\\"REPLACE\\\"), app)","start":STARTINGTIME,"step":60}}}},"strategy":"rollover"}\n'

# url = 'https://intuit.wavefront.com/api/v2/chart/metric/detail?m=appdynamics.apm.transactions.50th_percentile_resp_time_ms'
url = 'https://intuit.wavefront.com/api/v2/chart/api?q=avg(align(60s%2C%20mean%2C%20ts(appdynamics.apm.transactions.90th_percentile_resp_time_ms%2C%20env%3Dprd))%2C%20app)&s=1552341205&g=m&sorted=false&cached=true'

apitoken = {'Authorization': 'Bearer ba2e4e71-cf26-4e7c-afce-55b58446f991'}

r = requests.get(url,headers=apitoken)

type = "json" # "json"
# print(r.json()["hosts"])
# print(r.json()["hosts"])

i = 0
count = 0
# for host in r.json()["hosts"]:
    # if i > 100:
    #     count += 1
    #     i = 0
    # filename = ""
    # if type == "json":
    #     filename = "siminput/simrequests" + str(count) + ".txt"
    # else:
    #     filename = "siminput/simrequests" + str(count) + ".csv"
    # with open(filename, "a+") as f:
    #     starttime = random.randint(1550052000000, 1550055600000) #10 to 11 am
    #     endtime = random.randint(1550073600000, 1550077200000) #4 to 5 pm
    #     if type == "json":
    #         curtext = basetext
    #         curtext = curtext.replace('STARTINGTIME', str(starttime))
    #         curtext = curtext.replace('ENDINGTIME', str(endtime))
    #         curtext = curtext.replace('APP_CHANGE', host["host"])
    #         curtext = curtext.replace('REPLACE', host["tags"]["app"])
    #
    #         f.write(curtext)
    #
    #     else:
    #         line = ""
    #         line += host["tags"]["app"] + ","
    #         line += "4xx" + "," + "sum(ts(telegraf.http.server.requests.count, env=\\\"prd\\\" and app=\\\"REPLACE\\\" and status=\\\"4*\\\"))".replace("REPLACE", host["tags"]["app"]) + ","
    #         line += "5xx" + "," + "sum(ts(telegraf.http.server.requests.count, env=\\\"prd\\\" and app=\\\"REPLACE\\\" and status=\\\"5*\\\"))".replace("REPLACE", host["tags"]["app"]) + "\n"
    #         f.write(line)
    # i += 1

with open("appnames2.txt", "a+") as f:
    apps = {}
    for host in r.json()["timeseries"]:
        # apps[host["host"]] = True
        apps[host["tags"]["app"]] = True
    # print(apps)
    # print r.json()["timeseries"]
    for host in apps.keys():
        f.write(host+"\n")
