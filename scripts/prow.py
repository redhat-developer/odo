import pika, sys, os, json

# Parse CLODUAMQP_URL (fallback to localhost)
success = False
url = os.environ.get('CLOUDAMQP_URL', 'amqps://fjuitskq:WJ9hHhwhVXsc6N7J7lLQrEVFeIiDdRxY@shrimp.rmq.cloudamqp.com/fjuitskq')
params = pika.URLParameters(url)
params.socket_timeout = 5
jss = os.getenv('JOB_SPEC')
js = json.loads(jss)
pr_no = js['refs']['pulls'][0]['number']

connectionsend = pika.BlockingConnection(params) # Connect to CloudAMQP
channelsend = connectionsend.channel() # start a channel
channelsend.queue_declare(None, queue='prow_send', auto_delete=True) # Declare a queue
# send a message

channelsend.basic_publish(exchange='', routing_key='prow_send', body=str(pr_no))
print ("[x] Message sent to consumer, " + str(pr_no))
connectionsend.close()

# Now wait for response
params.blocked_connection_timeout = 2400
connectionrcv = pika.BlockingConnection(params) # Connect to CloudAMQP
channelrcv = connectionrcv.channel() # start a channel
channelrcv.queue_declare(None, queue='prow_rcv', auto_delete=True) # declare queue
channelrcv.exchange_declare(None, exchange=str(pr_no), exchange_type='topic', auto_delete=True) # declare exchange for pr_no
channelrcv.queue_bind(None, 'prow_rcv', str(pr_no)) # bind recieving queue to exchange

# Callback to operate on messege
def callback(ch, method, properties, body):
    print(" [x] Recieved " + body)
    data = json.loads(body)
    success = data['success']
    logs = data['logs']
    if not success:
        print(logs)
    ch.basic_ack(delivery_tag=method.delivery_tag)
    ch.exchange_delete(None, str(body))
    ch.stop_consuming()

channelrcv.basic_consume(callback, queue='prow_rcv')
channelrcv.start_consuming()
connectionrcv.close()
if not success:
    sys.exit(1)
