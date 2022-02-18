package logger

import "testing"

// go test -v -bench=New -benchtime=10s
//func BenchmarkNew(b *testing.B) {
//	logger := New()
//	logoTask := func() {
//		logger.Debug("debug...")
//		logger.Trace("trace...")
//		logger.Info("info...")
//		logger.Warning("warning...")
//		logger.Error("error...")
//	}
//
//	logger.Debug("sss")
//	// 开始性能测试
//	b.ReportAllocs()
//	b.StartTimer()
//	for i := 0; i < b.N; i++ {
//		logoTask()
//	}
//}

func TestName(t *testing.T) {
	log :=New()
	log.Info("test")
}