import pika, sys, os, json

# Parse CLODUAMQP_URL (fallback to localhost)
success = False
url = os.environ.get('CLOUDAMQP_URL', 'amqps://fjuitskq:WJ9hHhwhVXsc6N7J7lLQrEVFeIiDdRxY@shrimp.rmq.cloudamqp.com/fjuitskq')
params = pika.URLParameters(url)
params.socket_timeout = 5
jss = os.getenv('JOB_SPEC')
js = json.loads(jss)
pr_no = js['refs']['pulls'][0]['num']

connectionsend = pika.BlockingConnection(params) # Connect to CloudAMQP
channelsend = connectionsend.channel() # start a channel
channelsend.queue_declare(queue='prow_send') # Declare a queue
# send a message

channel.basic_publish(exchange='', routing_key='prow_send', body=str(pr_no))
print ("[x] Message sent to consumer, " + str(pr_no))
connectionsend.close()

# Now wait for response
connectionrcv = pika.BlockingConnection(params) # Connect to CloudAMQP
channelrcv = connectionrcv.channel() # start a channel
channelrcv.queue_declare('prow_rcv') # declare queue
channelrcv.exchange_declare(exchange=str(pr_no), exchange_type='topic') # declare exchange for pr_no
channelrcv.queue_bind(exchange=str(pr_no), queue='prow_rcv') # bind recieving queue to exchange

# Callback to operate on messege
def callback(ch, method, properties, body):
    print(" [x] %r:%r" % (method.routing_key, body))
    data = json.loads(body)
    success = data['success']
    logs = data['logs']
    if not success:
        print(logs)

channelrcv.basic_consume(queue='prow_rcv', on_message_callback=callback, auto_ack=True)
channelrcv.consume('prow_rcv', auto_ack=True, inactivity_timeout=2040)
channelrcv.exchange_delete(str(pr_no))
connectionrcv.close()
if not success:
    sys.exit(1)
