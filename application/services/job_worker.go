package services

import (
	"encoder/domain"
	"encoder/framework/utils"
	"encoding/json"
	"os"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/streadway/amqp"
)

type JobWorkerResult struct {
	Job     domain.Job
	Message *amqp.Delivery
	Error   error
}

var Mutex = &sync.Mutex{}

func JobWorker(messageChannel chan amqp.Delivery, returnChan chan JobWorkerResult, jobService JobService, job domain.Job, workerId int) {

	// {
	//	 "id": "525b5fd9-700d-4feb-89c0-415a1e6e148c",
	//	 "file_path": "convite.mp4",
	//	 "status": "pending"
	// }

	for message := range messageChannel {

		// pega message body do json
		err := utils.IsJson(string(message.Body))

		// validar se o json é um json
		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		Mutex.Lock()
		err = json.Unmarshal(message.Body, &jobService.VideoService.Video)
		jobService.VideoService.Video.ID = uuid.NewV4().String()
		Mutex.Unlock()

		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		// validar o vídeo
		err = jobService.VideoService.Video.Validade()

		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		Mutex.Lock()
		// inserir vídeo no banco de dados
		err = jobService.VideoService.InsertVideo()
		Mutex.Unlock()

		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		// start para iniciar o job
		job.Video = jobService.VideoService.Video
		job.OutputBucketPath = os.Getenv("outputBucketName")
		job.ID = uuid.NewV4().String()
		job.Status = "STARTING"
		job.CreatedAt = time.Now()

		Mutex.Lock()
		_, err = jobService.JobRepository.Insert(&job)
		Mutex.Unlock()

		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		jobService.Job = &job
		err = jobService.Start()

		if err != nil {
			returnChan <- returnJobResult(domain.Job{}, message, err)
			continue
		}

		returnChan <- returnJobResult(job, message, nil)

	}

}

func returnJobResult(job domain.Job, message amqp.Delivery, err error) JobWorkerResult {
	result := JobWorkerResult{
		Job:     job,
		Message: &message,
		Error:   err,
	}

	return result
}
