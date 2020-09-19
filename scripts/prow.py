import pika, sys, os, time, json

def main():
    params = pika.URLParameters('amqps://fjuitskq:WJ9hHhwhVXsc6N7J7lLQrEVFeIiDdRxY@shrimp.rmq.cloudamqp.com/fjuitskq')
    params.socket_timeout = 6
    connections = pika.BlockingConnection(params)
    channels = connections.channel()
    channels.queue_declare(queue='prow_recieve')
    jss = os.getenv("JOB_SPEC")
    js = json.loads(jss)
    pr_no = js['refs']['pulls'][0]['num']
    body = "Please process PR " + pr_no
    channels.basic_publish(exchange='', routing_key='odo', body="")
    print(' [*] Sent ' + body)
    connections.close()
