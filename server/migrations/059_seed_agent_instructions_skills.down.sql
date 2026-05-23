-- Reverse: clear instructions, remove skills and agent_skill links for SDR agents
DO $$
DECLARE
    ws RECORD;
BEGIN
    FOR ws IN SELECT id FROM workspace LOOP
        UPDATE agent SET instructions = ''
        WHERE workspace_id = ws.id AND name IN ('Hunter', 'Qualificador', 'Copywriter', 'Closer', 'Nurture');

        DELETE FROM agent_skill WHERE agent_id IN (
            SELECT id FROM agent WHERE workspace_id = ws.id AND name IN ('Hunter', 'Qualificador', 'Copywriter', 'Closer', 'Nurture')
        );

        DELETE FROM skill_file WHERE skill_id IN (
            SELECT id FROM skill WHERE workspace_id = ws.id AND name IN ('Hunter-Prospector', 'Lead-Qualification', 'Sales-Copywriting', 'Deal-Closing', 'Nurture-Email')
        );

        DELETE FROM skill WHERE workspace_id = ws.id AND name IN ('Hunter-Prospector', 'Lead-Qualification', 'Sales-Copywriting', 'Deal-Closing', 'Nurture-Email');
    END LOOP;
END $$;