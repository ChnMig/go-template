package health

// StatusDTO 健康检查接口的返回数据结构
// 通过 DTO 与内部实现解耦，避免直接暴露内部模型。
type StatusDTO struct {
	Status    string `json:"status"`    // 服务整体健康状态
	Ready     bool   `json:"ready"`     // 是否就绪可对外提供服务
	Uptime    string `json:"uptime"`    // 服务运行时长（人类可读）
	Timestamp int64  `json:"timestamp"` // 当前时间戳（秒）
}
