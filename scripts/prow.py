import pika, sys, os, json, time

# Parse CLODUAMQP_URL (fallback to localhost)
success = False
url = os.environ.get('CLOUDAMQP_URL', 'amqps://fjuitskq:WJ9hHhwhVXsc6N7J7lLQrEVFeIiDdRxY@shrimp.rmq.cloudamqp.com/fjuitskq')
params = pika.URLParameters(url)
params.socket_timeout = 5
#jss = os.getenv('JOB_SPEC')
#js = json.loads(jss)
#prn = js['refs']['pulls'][0]['number']
#pr_no="{}".format(prn)
pr_no = '2521'
rcv_queue = "prow_recieve_{}".format(pr_no)

connectionsend = pika.BlockingConnection(params) # Connect to CloudAMQP
channelsend = connectionsend.channel() # start a channel
channelsend.queue_declare(queue='prow_send') # Declare a queue
# send a message

channelsend.basic_publish(exchange='', routing_key='prow_send', body=pr_no)
print (" [x] Message sent to consumer, " + pr_no)
connectionsend.close()

# Now wait for response
params.blocked_connection_timeout = 2400
connectionrcv = pika.BlockingConnection(params) # Connect to CloudAMQP
channelrcv = connectionrcv.channel() # start a channel
channelrcv.queue_declare(queue=rcv_queue)

# Callback to operate on messege
def callback(ch, method, properties, body):
    data = json.loads(body)
    if data['kind'] != "status":
        print(data['data'])
        ch.basic_ack(delivery_tag=method.delivery_tag)
    else:
        success = data['data']
        if not success:
            print(" [x] FAIL !! See logs above  " )
        ch.basic_ack(delivery_tag=method.delivery_tag)
        ch.stop_consuming()

channelrcv.basic_consume(callback, queue=rcv_queue)
print(" [x] starting consumption on  " + rcv_queue)
channelrcv.start_consuming()
channelrcv.queue_delete(queue=rcv_queue)
connectionrcv.close()
if not success:
    sys.exit(1)
