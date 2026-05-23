-- Migrate to the 5-agent SOTA architecture (event-driven state machine).
-- Removes intermediate/legacy agents and ensures only Hunter, Qualificador, Copywriter, Closer, Nurture exist.

DO $$
DECLARE
    ws RECORD;
    default_runtime UUID;
BEGIN
    FOR ws IN SELECT id FROM workspace LOOP
        SELECT id INTO default_runtime FROM agent_runtime WHERE workspace_id = ws.id ORDER BY created_at LIMIT 1;
        IF default_runtime IS NULL THEN
            INSERT INTO agent_runtime (workspace_id, daemon_id, name, runtime_mode, provider, status)
            VALUES (ws.id, 'cloud-daemon', 'Default Runtime', 'cloud', 'multica', 'active')
            RETURNING id INTO default_runtime;
        END IF;
        -- Rename Prospector → Hunter (avoids duplicate)
        UPDATE agent SET name = 'Hunter'
        WHERE workspace_id = ws.id AND name = 'Prospector'
          AND NOT EXISTS (SELECT 1 FROM agent WHERE workspace_id = ws.id AND name = 'Hunter');

        -- Rename Mailer → Nurture (avoids duplicate)
        UPDATE agent SET name = 'Nurture'
        WHERE workspace_id = ws.id AND name = 'Mailer'
          AND NOT EXISTS (SELECT 1 FROM agent WHERE workspace_id = ws.id AND name = 'Nurture');

        -- Remove intermediate agents no longer part of the 5-agent design
        DELETE FROM agent
        WHERE workspace_id = ws.id AND name IN ('Investigator', 'SDR', 'Prospector', 'Mailer')
          AND NOT EXISTS (SELECT 1 FROM agent_task_queue WHERE agent_id = agent.id);

        -- Hunter: finds and enriches leads
        INSERT INTO agent (workspace_id, name, description, runtime_mode, runtime_config, visibility, status, max_concurrent_tasks, instructions, custom_env, custom_args, handoff_mode, runtime_id)
        SELECT ws.id, 'Hunter', 'Caçador de leads B2B. Busca, valida ICP e enriquece dados de leads em Apollo.io, CSV ou inbound forms. Qualidade > quantidade.', 'cloud', '{}', 'workspace', 'offline', 6, '', '{}', '[]', 'automatic', default_runtime
        WHERE NOT EXISTS (SELECT 1 FROM agent WHERE workspace_id = ws.id AND name = 'Hunter');

        -- Qualificador: scores and qualifies leads
        INSERT INTO agent (workspace_id, name, description, runtime_mode, runtime_config, visibility, status, max_concurrent_tasks, instructions, custom_env, custom_args, handoff_mode, runtime_id)
        SELECT ws.id, 'Qualificador', 'Qualificador de Leads Senior. Aplica rubrica BANT+IA e decide se lead deve ser descartado, nutrido ou encaminhado.', 'cloud', '{}', 'workspace', 'offline', 6, '', '{}', '[]', 'automatic', default_runtime
        WHERE NOT EXISTS (SELECT 1 FROM agent WHERE workspace_id = ws.id AND name = 'Qualificador');

        -- Copywriter: generates personalized email copy
        INSERT INTO agent (workspace_id, name, description, runtime_mode, runtime_config, visibility, status, max_concurrent_tasks, instructions, custom_env, custom_args, handoff_mode, runtime_id)
        SELECT ws.id, 'Copywriter', 'Copywriter de alta conversão. Escreve copy que vende sem ser apelativo, adaptada ao perfil do lead.', 'cloud', '{}', 'workspace', 'offline', 6, '', '{}', '[]', 'automatic', default_runtime
        WHERE NOT EXISTS (SELECT 1 FROM agent WHERE workspace_id = ws.id AND name = 'Copywriter');

        -- Closer: defines outreach strategy and handles objections
        INSERT INTO agent (workspace_id, name, description, runtime_mode, runtime_config, visibility, status, max_concurrent_tasks, instructions, custom_env, custom_args, handoff_mode, runtime_id)
        SELECT ws.id, 'Closer', 'Closer estrategista. Prepara abordagem, rebatimento de objeções e decide se converte solo ou passa para humano.', 'cloud', '{}', 'workspace', 'offline', 6, '', '{}', '[]', 'automatic', default_runtime
        WHERE NOT EXISTS (SELECT 1 FROM agent WHERE workspace_id = ws.id AND name = 'Closer');

        -- Nurture: sends emails, monitors events, adaptive follow-up, handoff to human
        INSERT INTO agent (workspace_id, name, description, runtime_mode, runtime_config, visibility, status, max_concurrent_tasks, instructions, custom_env, custom_args, handoff_mode, runtime_id)
        SELECT ws.id, 'Nurture', 'Nutridor inteligente do pipeline. Monitora eventos de email, adapta follow-up em tempo real e decide o momento exato do handoff para humano.', 'cloud', '{}', 'workspace', 'offline', 6, '', '{}', '[]', 'automatic', default_runtime
        WHERE NOT EXISTS (SELECT 1 FROM agent WHERE workspace_id = ws.id AND name = 'Nurture');
    END LOOP;
END $$;
