package queue

import (
	"github.com/iVariable/qp/src/qp"
	"github.com/iVariable/qp/src/utils"

	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// Sqs - AWS SQS implementation
type Sqs struct {
	configuration sqsConfiguration
	queue         *sqs.SQS
	queueURL      *string
	logger        *log.Entry
}

type sqsConfiguration struct {
	QueueName       string
	WaitTimeSeconds int
	AwsRegion       string
	AwsProfile      string
}

// Configure configure queue
func (q *Sqs) Configure(configuration map[string]interface{}) error {
	q.logger = log.WithFields(log.Fields{
		"type":  "queue",
		"queue": "Sqs",
	})
	q.logger.Debug("Reading configuration")
	q.configuration.WaitTimeSeconds = 20 //Defaults
	utils.FillStruct(configuration, &q.configuration)

	if q.configuration.AwsProfile == "" {
		return errors.New("You need to provide AwsProfile for Sqs queue")
	}

	if q.configuration.AwsRegion == "" {
		return errors.New("You need to provide AwsRegion for Sqs queue")
	}

	q.queue = sqs.New(session.New(), &aws.Config{
		Region:      aws.String(q.configuration.AwsRegion),
		Credentials: credentials.NewSharedCredentials("", q.configuration.AwsProfile),
	})

	params := &sqs.GetQueueUrlInput{
		QueueName: aws.String(q.configuration.QueueName),
	}

	resp, err := q.queue.GetQueueUrl(params)

	if err != nil {
		q.logger.WithError(err).Error("Error on GetQueueUrl")
		return err
	}

	q.queueURL = resp.QueueUrl

	q.logger.WithField("configuration", q.configuration).Info("Configuration loaded")

	return nil
}

// GetName returns name of the queue
func (q *Sqs) GetName() string {
	return "Sqs"
}

// Consume consume a message from the queue
func (q *Sqs) Consume() (qp.IMessage, error) {
	q.logger.Debug("Message consume")
	for {
		params := &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(*q.queueURL),
			MaxNumberOfMessages: aws.Int64(1),
			WaitTimeSeconds:     aws.Int64(int64(q.configuration.WaitTimeSeconds)),
		}

		resp, err := q.queue.ReceiveMessage(params)

		if err != nil {
			return nil, err
		}

		if len(resp.Messages) != 0 {
			return &qp.Message{
				ID:   *resp.Messages[0].ReceiptHandle,
				Body: *resp.Messages[0].Body,
				Raw:  resp.GoString(),
			}, nil
		}
	}
}

// Ack acknowledge a message
func (q *Sqs) Ack(message qp.IMessage) error {
	q.logger.WithField("message", message).Debug("Message acknowledged")
	params := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(*q.queueURL),
		ReceiptHandle: aws.String((message.GetID()).(string)),
	}

	_, err := q.queue.DeleteMessage(params)

	if err != nil {
		return err
	}

	return nil
}

// Reject reject a message
func (q *Sqs) Reject(message qp.IMessage) error {
	q.logger.WithField("message", message).Debug("Message rejected")
	// Do nothing. Aws SQS will take care of not acknowledged messages
	// and will put them into dead letter queue for us
	// We can integrate here some RejectPolicy, but do not want to spent time on this now
	return nil
}

// GetNumberOfMessages returns the number of messages
func (q *Sqs) GetNumberOfMessages() (int, error) {
	return 9999999, nil
}
