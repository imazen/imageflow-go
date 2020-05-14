package imageflow

import (
	"io/ioutil"
	"testing"
)

var data []byte

func init() {
	data, _ = ioutil.ReadFile("image.jpg")
}

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
	step.Decode(&File{filename: "image.jpg"}).FlipV().Watermark(&File{filename: "image.jpg"}, nil, "within", PercentageFitBox{
		X1: 0,
		Y1: 0,
		X2: 50,
		Y2: 50,
	}, 0.3, nil).ConstrainWithinW(400).FillRect(0, 0, 8, 8, Black{}).Branch(func(step *Steps) {
		step.ConstrainWithin(100, 100).Rotate180().Region(Region{
			X1:              0,
			Y1:              0,
			X2:              200,
			Y2:              200,
			BackgroundColor: Black{},
		}).Branch(func(step *Steps) {
			step.GrayscaleFlat().Encode(&File{filename: "gray_small.jpg"}, MozJPEG{})
		}).Encode(&File{filename: "small.jpg"}, MozJPEG{})
	}).ExpandCanvas(ExpandCanvas{Top: 10, Color: Black{}}).Encode(&File{filename: "medium.jpg"}, MozJPEG{}).Execute()
}

func BenchmarkSteps(b *testing.B) {
	for i := 0; i < b.N; i++ {
		step := NewStep()
		step.Decode(&Buffer{buffer: data}).ConstrainWithinW(400).Branch(func(step *Steps) {
			step.ConstrainWithin(100, 100).Region(Region{
				X1:              0,
				Y1:              0,
				X2:              200,
				Y2:              200,
				BackgroundColor: Black{},
			}).Rotate180().Encode(&Buffer{}, MozJPEG{})
		}).Encode(&Buffer{}, MozJPEG{}).Execute()
	}
}
