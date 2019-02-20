-- +goose Up
WITH moduleVar AS (
    INSERT INTO modules (instance_id, name, active) VALUES (1, 'mdm-converter', 't') RETURNING id
)

INSERT INTO configs (module_id, active, data) VALUES (
  (SELECT id FROM moduleVar),
  't',
  ('{
  "rules": {
    "1": {
      "availableFields": [
        "*"
      ],
      "formatting": {
        "sudir_v1": {
          "birth_dt": "dateFormatter",
          "documents.1000016.start_dt": "dateFormatter"
        },
        "sudir_v2": {
          "birth_dt": "dateFormatter",
          "documents.1000016.start_dt": "dateFormatter"
        },
        "sudir_v1_find": {
          "documents.1000016.start_dt": "dateFormatter",
          "birth_dt": "dateFormatter"
        },
        "sudir_v2_find": {
          "birth_dt": "dateFormatter",
          "documents.1000016.start_dt": "dateFormatter"
        }
      },
      "defaultValues": {
        "sma": {
          "*": "%undefined%"
        }
      }
    },
    "9": {
      "availableFields": [
        "*"
      ]
    },
    "10": {
      "availableFields": [
        "*"
      ]
    }
  },
  "metrics": {
    "address": {
      "ip": "0.0.0.0",
      "path": "/metrics",
      "port": "9571"
    },
    "gc": true,
    "memory": true
  },
  "formatters": {
    "dateFormatter": {
      "type": "DATE_TIME",
      "outFormat": "2006-01-02",
      "inputFormat": "2006-01-02 15:04:05.999"
    }
  },
  "mappingSchema": {
    "sudir_v2_find": {
      "type": "sudir",
      "extends": [
        "sudir_v2"
      ]
    },
    "erl": {
      "extends": [],
      "mapping": {}
    },
    "sma": {
      "mapping": {
      }
    },
    "sudir": {
      "type": "sudir",
      "mapping": {
      }
    },
    "sudir_v1": {
      "type": "sudir",
      "extends": [
        "sudir"
      ],
      "mapping": {}
    },
    "sudir_v2": {
      "type": "sudir",
      "extends": [
        "sudir"
      ],
      "mapping": {}
    },
    "sudir_v1_find": {
      "type": "sudir",
      "extends": [
        "sudir_v1"
      ]
    },
    "test_protocol": {
      "mapping": {
        "etalon_id": "id"
      }
    }
  }
}
'
)::jsonb);

-- +goose Down

