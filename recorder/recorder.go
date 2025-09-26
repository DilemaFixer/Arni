package recorder

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/gen2brain/malgo"
)

type VoiceRecorder struct {
	ctx            *malgo.AllocatedContext
	captureDevice  *malgo.Device
	playbackDevice *malgo.Device
	AudioBuffer    []float32
	isRecording    bool
}

func NewVoiceRecorder() (*VoiceRecorder, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		// log.Printf("Malgo: %s", message)
	})
	if err != nil {
		return nil, err
	}

	return &VoiceRecorder{
		ctx:         ctx,
		AudioBuffer: make([]float32, 0),
		isRecording: false,
	}, nil
}

func (vr *VoiceRecorder) StartRecording() error {
	if vr.isRecording {
		return fmt.Errorf("recording is tarted before")
	}

	vr.AudioBuffer = vr.AudioBuffer[:0]

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatF32
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 16000

	dataCallback := func(pOutputSample, pInputSample []byte, framecount uint32) {
		if !vr.isRecording {
			return
		}
		sampleCount := framecount * deviceConfig.Capture.Channels
		for i := uint32(0); i < sampleCount; i++ {
			offset := i * 4 
			if int(offset+4) <= len(pInputSample) {
				sample := *(*float32)(unsafe.Pointer(&pInputSample[offset]))
				vr.AudioBuffer = append(vr.AudioBuffer, sample)
			}
		}
	}

	device, err := malgo.InitDevice(vr.ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: dataCallback,
	})
	if err != nil {
		return err
	}

	vr.captureDevice = device
	vr.isRecording = true

	return device.Start()
}

func (vr *VoiceRecorder) StopRecording() error {
	if !vr.isRecording {
		return fmt.Errorf("recoding is not started")
	}

	vr.isRecording = false

	if vr.captureDevice != nil {
		_ = vr.captureDevice.Stop()
		vr.captureDevice.Uninit()
		vr.captureDevice = nil
	}

	return nil
}

func (vr *VoiceRecorder) PlayRecording() error {
	if len(vr.AudioBuffer) == 0 {
		return fmt.Errorf("not data")
	}

	samples := make([]float32, len(vr.AudioBuffer))
	copy(samples, vr.AudioBuffer)

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatF32
	deviceConfig.Playback.Channels = 1
	deviceConfig.SampleRate = 16000

	currentSample := uint32(0)
	totalSamples := uint32(len(samples))

	playCallback := func(pOutputSample, pInputSample []byte, framecount uint32) {
		sampleCount := framecount * deviceConfig.Playback.Channels
		for i := uint32(0); i < sampleCount; i++ {
			offset := i * 4
			if int(offset+4) <= len(pOutputSample) {
				var sample float32
				if currentSample < totalSamples {
					sample = samples[currentSample]
					currentSample++
				} else {
					sample = 0
				}
				*(*float32)(unsafe.Pointer(&pOutputSample[offset])) = sample
			}
		}
	}

	device, err := malgo.InitDevice(vr.ctx.Context, deviceConfig, malgo.DeviceCallbacks{
		Data: playCallback,
	})
	if err != nil {
		return err
	}

	vr.playbackDevice = device

	if err := device.Start(); err != nil {
		return err
	}

	duration := time.Duration(len(samples)) * time.Second / 16000
	time.Sleep(duration + 200*time.Millisecond)

	_ = device.Stop()
	device.Uninit()
	vr.playbackDevice = nil

	return nil
}

func (vr *VoiceRecorder) Close() {
	if vr.captureDevice != nil {
		vr.captureDevice.Uninit()
		vr.captureDevice = nil
	}
	if vr.playbackDevice != nil {
		vr.playbackDevice.Uninit()
		vr.playbackDevice = nil
	}
	if vr.ctx != nil {
		vr.ctx.Uninit()
		vr.ctx.Free()
	}
}
