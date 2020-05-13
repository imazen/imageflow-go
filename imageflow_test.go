package imageflow

import (
	"io/ioutil"
	"testing"
)

func TestJob(t *testing.T) {
	job := New()
	data, _ := ioutil.ReadFile("image.jpg")
	command := []byte("{\"io\":[{\"io_id\":0,\"direction\":\"in\",\"io\":\"placeholder\"},{\"io_id\":1,\"direction\":\"out\",\"io\":\"placeholder\"}],\"framewise\":{\"steps\":[{\"decode\":{\"io_id\":0}},{\"constrain\":{\"mode\":\"within\",\"w\":400}},\"rotate_90\",{\"encode\":{\"io_id\":1,\"preset\":{\"pngquant\":{\"quality\":80}}}}]}}")
	job.AddInput(0, data)
	job.AddOutput(1)
	job.Message(command)
	ioutil.WriteFile("./output.jpg", job.GetOutput(1), 0644)
}

func TestStep(t *testing.T) {
	step := NewStep()
	step.Decode(&File{filename: "image.jpg"}).ConstrainWithinW(400).Branch(func(step *Steps) {
		step.ConstrainWithin(100, 100).Encode(&File{filename: "small.jpg"}, MozJPEG{})
	}).Encode(&File{filename: "medium.jpg"}, MozJPEG{}).Execute()
}
