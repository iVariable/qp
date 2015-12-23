# qp (Queue Processor)

Qp is a queue processor which aims for several goals:

- to abstract out trivial-mechanic queue processing actions (pull message from the queue, ack, reject, etc) from your application business logic.
- to separate queue implementation/details from actual message processing logic

Those goals are achieved thru separation of Queue Consumers and Queue Processors.

Ordinary implementation of queue processor is a daemon which pulls messages from the queue, does some actions (business logic) 
and then removes message (acknowledges or rejects) from the queue. 

But what to do if your business application language or application itself does not support daemon-mode? 

Or your application can handle only limited amount of messages per second?

Or your application does requests to external resources which has rate-limiting (Github API, Google Analytics API, etc)?

Of course you can code all those cases in your app, rewrite it or do something else :) But do you want to spend time for this?

qp to the rescue!

# Examples

## SQS to HTTP

Queue: AWS SQS

QueueProcessor: HTTPProxy

With config below you will be consuming SQS in 10 parallel threads. All messages from the queue will redirected to ```http://my-api.com/api/v1/resizeImage/``` endpoint.

If endpoint returns 200 message considered acknowledged and removed from the queue. Any other response is considered as failed and message is rejected (it stays in the queue until it moved to dead-letters queue by AWS SQS).

    strategy:
      - name: Processing without limits
        type: ParallelProcessing
        options:
          MaxThreads: 10
          OnProcessingError: warning
          Queue: Images queue
          Processor: Image resizer
    
    queue:
      - name: Images queue
        type: Sqs
        options:
          QueueName: "resizeImage"
          AwsRegion: eu-west-1
          AwsProfile: mycoolapp
          WaitTimeSeconds: 20
    
    processor:
      - name: Image resizer
        type: HTTPProxy
        options:
          Timeout: 5
          URL: "http://my-api.com/api/v1/resizeImage/"

## SQS to Custom local script/program

Queue: AWS SQS

QueueProcessor: Custom local script/program

With config below you will be consuming SQS in 1 threads with limit not more than 17 messages per second (can be less if you processor takes time). 
All messages from the queue will be send to bash script you define. If script exists with 0 exit code message considered successfully processed. Else - message is rejected.

    strategy:
      - name: Convert pdf
        type: ParallelProcessing
        options:
          MaxThreads: 1
          ProcessorThroughput: 17
          OnProcessingError: warning
          Queue: Doc to pdf
          Processor: HttpProxy
    
    queue:
      - name: Sqs
        type: Sqs
        options:
          QueueName: "doc-to-pdf"
          AwsRegion: eu-west-1
          AwsProfile: mycoolconverter
          WaitTimeSeconds: 20
    
    processor:
      - name: Doc to pdf
        type: Shell
        options:
          MessagePlaceholder: "%msg%"
          Command: "doc-to-pdf.sh %msg%"
          EchoOutput: false

# Supported processing strategies

## ParallelProcessing

This strategy consumes one queue and redirects messages to one processor in multiple threads with rate-limiting capabilities

# Supported Queues

## AWS SQS 

https://aws.amazon.com/sqs/
      
## Tail

Pseudo-queue. Tails file. Each line treated as a message. Can be used to send local logs to central storage, for example.

## Dummy 

Pseudo-queue. Generates messages with randomized delay in between. Useful for debugging

# Supported Processors

## HTTPProxy

Proxies messages to any HTTP endpoint.
Each message is sent via POST request. Message body holds actual queue message.

If endpoint returns 200 message considered acknowledged and removed from the queue. 

Any other response is considered as failed and message is rejected.

## Shell 

Proxies message to shell script.

If script exists with 0 exit code message considered successfully processed. Else - message is rejected.

## Stdout

Outputs message to stdout. Useful for debugging