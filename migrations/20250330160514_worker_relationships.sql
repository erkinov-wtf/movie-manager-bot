-- Modify "worker_tasks" table
ALTER TABLE "worker_tasks" ADD CONSTRAINT "fk_worker_tasks_worker_id" FOREIGN KEY ("worker_id") REFERENCES "worker_states" ("worker_id") ON UPDATE NO ACTION ON DELETE NO ACTION;
