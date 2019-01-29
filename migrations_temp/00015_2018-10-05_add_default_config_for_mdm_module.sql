-- +goose Up
set migrationRead.db_host = 'msk-pgsql.isp.mobi';
set migrationRead.db_port = '5432';
set migrationRead.db_dbname = 'isp-alpha';
set migrationRead.db_user = 'isp-alpha';
set migrationRead.db_pass = 'JxY96xLF8zfh3g8EWvSDtLUBw';

set migrationWrite.db_host = 'msk-pgsql.isp.mobi';
set migrationWrite.db_port = '5432';
set migrationWrite.db_dbname = 'isp-alpha';
set migrationWrite.db_user = 'isp-alpha';
set migrationWrite.db_pass = 'JxY96xLF8zfh3g8EWvSDtLUBw';

set migrationNotification.db_host = 'msk-pgsql.isp.mobi';
set migrationNotification.db_port = '5432';
set migrationNotification.db_dbname = 'isp-alpha';
set migrationNotification.db_user = 'isp-alpha';
set migrationNotification.db_pass = 'JxY96xLF8zfh3g8EWvSDtLUBw';

WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'mdm', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ('{"databaseRead": {' ||
   '"port": "' || current_setting('migrationRead.db_port') || '", ' ||
    '"schema": "mdm_service", ' ||
     '"address": "' || current_setting('migrationRead.db_host') || '", ' ||
      '"database": "' || current_setting('migrationRead.db_dbname') || '", ' ||
       '"password": "' || current_setting('migrationRead.db_pass') || '", ' ||
        '"username": "' || current_setting('migrationRead.db_user') || '", ' ||
         '"createSchema": true},' ||
          '"metrics": {"gc": true, "memory": true, "address": {"ip": "0.0.0.0", "path": "/metrics", "port": "9568"}},' ||
           '"databaseWrite": {' ||
            '"port": "' || current_setting('migrationWrite.db_port') || '", ' ||
             '"schema": "mdm_service", ' ||
              '"address": "' || current_setting('migrationWrite.db_host') || '", ' ||
               '"database": "' || current_setting('migrationWrite.db_dbname') || '", ' ||
                '"password": "' || current_setting('migrationWrite.db_pass') || '", ' ||
                 '"username": "' || current_setting('migrationWrite.db_user') || '", ' ||
                  '"createSchema": true},' ||
                   '"databaseNotifications": {' ||
                    '"port": "' || current_setting('migrationNotification.db_port') || '", ' ||
                     '"schema": "mdm_service", ' ||
                      '"address": "' || current_setting('migrationNotification.db_host') || '", ' ||
                       '"database": "' || current_setting('migrationNotification.db_dbname') || '", ' ||
                        '"password": "' || current_setting('migrationNotification.db_pass') || '", ' ||
                         '"username": "' || current_setting('migrationNotification.db_user') || '", ' ||
                          '"createSchema": true}}')::jsonb);

-- +goose Down

