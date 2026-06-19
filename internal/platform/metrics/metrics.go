package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// All counters/histograms/gauges from the architecture spec's observability
// table (§5.8), registered once at package init via promauto and read by
// every package that needs to record something -- there's deliberately no
// metrics interface/mock here, since Prometheus client vars are already
// trivial to no-op in tests by just not asserting on them.
var (
	LoginSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rbac_login_success_total",
		Help: "Total successful logins.",
	})
	LoginFailureTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rbac_login_failure_total",
		Help: "Total failed login attempts.",
	})
	UsersCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rbac_users_created_total",
		Help: "Total users created via signup.",
	})
	RoleAssignmentsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rbac_role_assignments_total",
		Help: "Total role assignments.",
	})
	PermissionChecksTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rbac_permission_checks_total",
		Help: "Total /authorize checks, labeled by cache hit/miss and allow/deny outcome.",
	}, []string{"cache", "result"})
	RedisCacheHitTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rbac_redis_cache_hit_total",
		Help: "Total permission cache hits.",
	})
	RedisCacheMissTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rbac_redis_cache_miss_total",
		Help: "Total permission cache misses (includes Redis being unreachable).",
	})
	RabbitMQPublishedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rbac_rabbitmq_messages_published_total",
		Help: "Total messages the outbox relay published to RabbitMQ, by event type.",
	}, []string{"event_type"})
	EventsConsumedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rbac_events_consumed_total",
		Help: "Total events consumed, by event type and consumer.",
	}, []string{"event_type", "consumer"})
	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "rbac_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path", "status"})
	OutboxRelayLagSeconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "rbac_outbox_relay_lag_seconds",
		Help: "Age in seconds of the oldest pending outbox row.",
	})
)
