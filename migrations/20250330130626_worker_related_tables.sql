-- Create "worker_states" table
CREATE TABLE "worker_states" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "worker_id" character varying(255) NOT NULL,
  "worker_type" character varying(50) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'idle',
  "last_check_time" timestamptz NULL,
  "next_check_time" timestamptz NULL,
  "error" text NULL,
  "shows_checked" integer NOT NULL DEFAULT 0,
  "updates_found" integer NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id"),
  CONSTRAINT "worker_states_worker_id_key" UNIQUE ("worker_id")
);
-- Create index "idx_worker_states_worker_id" to table: "worker_states"
CREATE INDEX "idx_worker_states_worker_id" ON "worker_states" ("worker_id");
-- Create "worker_tasks" table
CREATE TABLE "worker_tasks" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "worker_id" character varying(255) NOT NULL,
  "task_type" character varying(50) NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'running',
  "start_time" timestamptz NOT NULL,
  "end_time" timestamptz NULL,
  "duration_ms" bigint NULL,
  "error" text NULL,
  "show_id" bigint NULL,
  "user_id" bigint NULL,
  "shows_checked" integer NULL,
  "updates_found" integer NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY ("id")
);
-- Create index "idx_worker_tasks_created_at" to table: "worker_tasks"
CREATE INDEX "idx_worker_tasks_created_at" ON "worker_tasks" ("created_at");
-- Create index "idx_worker_tasks_status" to table: "worker_tasks"
CREATE INDEX "idx_worker_tasks_status" ON "worker_tasks" ("status");
-- Create index "idx_worker_tasks_worker_id" to table: "worker_tasks"
CREATE INDEX "idx_worker_tasks_worker_id" ON "worker_tasks" ("worker_id");
-- Create "worker_performance" view
CREATE VIEW "worker_performance" (
  "worker_id",
  "worker_type",
  "status",
  "last_check_time",
  "next_check_time",
  "shows_checked",
  "updates_found",
  "error",
  "total_tasks",
  "avg_task_duration_ms",
  "max_task_duration_ms",
  "error_tasks",
  "last_task_time"
) AS SELECT w.worker_id,
    w.worker_type,
    w.status,
    w.last_check_time,
    w.next_check_time,
    w.shows_checked,
    w.updates_found,
    w.error,
    count(t.id) AS total_tasks,
    avg(t.duration_ms) AS avg_task_duration_ms,
    max(t.duration_ms) AS max_task_duration_ms,
    sum(
        CASE
            WHEN ((t.status)::text = 'error'::text) THEN 1
            ELSE 0
        END) AS error_tasks,
    max(t.created_at) AS last_task_time
   FROM (worker_states w
     LEFT JOIN worker_tasks t ON (((w.worker_id)::text = (t.worker_id)::text)))
  GROUP BY w.id, w.worker_id, w.worker_type, w.status, w.last_check_time, w.next_check_time, w.shows_checked, w.updates_found, w.error;
