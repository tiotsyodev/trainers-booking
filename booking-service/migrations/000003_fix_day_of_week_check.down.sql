ALTER TABLE working_hours DROP CONSTRAINT working_hours_day_of_week_check;
ALTER TABLE working_hours ADD CONSTRAINT working_hours_day_of_week_check CHECK (day_of_week BETWEEN 0 AND 7);