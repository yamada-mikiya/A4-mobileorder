CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW(); -- 'NEW'は更新後の行データを表す
   RETURN NEW;
END;
$$ language 'plpgsql';