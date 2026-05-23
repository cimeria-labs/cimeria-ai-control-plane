-- Remove seeded sales agents only if they have no task history (safety guard).
DELETE FROM agent
WHERE name IN ('Hunter', 'Qualificador', 'Copywriter', 'Closer', 'Nurture')
  AND NOT EXISTS (SELECT 1 FROM agent_task_queue WHERE agent_id = agent.id);
