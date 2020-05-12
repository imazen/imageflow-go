package imageflow

import (
	"io/ioutil"
	"testing"
)

func TestJob(t *testing.T) {
	job := New()
	data, _ := ioutil.ReadFile("hello.jpg")
	commad := []byte("{\"io\":[{\"io_id\":0,\"direction\":\"in\",\"io\":\"placeholder\"},{\"io_id\":1,\"direction\":\"out\",\"io\":\"placeholder\"}],\"framewise\":{\"steps\":[{\"decode\":{\"io_id\":0}},{\"constrain\":{\"mode\":\"within\",\"w\":400}},\"rotate_90\",{\"encode\":{\"io_id\":1,\"preset\":{\"pngquant\":{\"quality\":80}}}}]}}")
	job.AddInput(0, data)
	job.AddOutput(1)
	job.Message(commad)
	ioutil.WriteFile("./ouput.jpg", job.GetOutput(1), 0644)
}
