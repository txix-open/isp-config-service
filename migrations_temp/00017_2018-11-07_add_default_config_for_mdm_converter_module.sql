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
            "etalon_id",
            "last_name",
            "$$documents.1000016",
            "vehicles",
            "addresses",
            "birth_dt",
            "gender_tp_code"
          ],
          "protocol": "sudir"
        },
        "2": {
          "protocol": "url",
          "availableFields": [
            "*"
          ]
        }
      },
      "metrics": {
        "gc": true,
        "memory": true,
        "address": {
          "ip": "0.0.0.0",
          "path": "/metrics",
          "port": "9571"
        }
      },
      "mappingSchema": {
        "url": {},
        "sudir": {
          "$$documents.1000019.ref_num": "DRIVER_LICENSE.DriverLicense",
          "gkupays.mesmeter": "[]HOUSE.MesMeterNumber",
          "$$documents.1000016.start_dt": "PASSPORT_RF.PassportIssueDate",
          "$$addresses.1000009.moscow_district_id": "[]REG_ADDRESS.RegAddressDistrictId",
          "$$addresses.7.unom": "[]ADDRESS.RealAddressUnom",
          "$$addresses.1000009.unom": "[]REG_ADDRESS.RegAddressUnom",
          "$$addresses.1000009.moscow_area_ext_id": "[]REG_ADDRESS.RegAddressAreaId",
          "$$addresses.7.moscow_area_ext_id": "[]ADDRESS.RealAddressAreaId",
          "$$escredentials.1000004.etalon_id": "[]OLYMPIAD.OlympiadId",
          "last_name": "FIO.LastName",
          "vehicles.reg_number": "[]VEHICLE.VehicleRegistrationNumber",
          "$$contacts.1.ref_num": "REG_DATA.Phone",
          "$$addresses.7.corpus_no": "[]ADDRESS.RealAddressCorpus",
          "$$users.NOT_SSO.id_value": "USER.LoginName",
          "$$addresses.7.residence_num": "[]ADDRESS.RealAddressFlat",
          "$$users.SSO.id_value": "USER.LoginName",
          "birth_place_line_one": "PASSPORT_RF.PassportBirthPlace",
          "$$addresses.7.house_no": "[]ADDRESS.RealAddressHouse",
          "$$addresses.1000009.street_name_ex": "[]REG_ADDRESS.RegAddressStreet",
          "gkupays.mesaccount": "[]HOUSE.MesAccountNumber",
          "vehicles.etalon_id": "[]VEHICLE.VehicleId",
          "$$addresses.7.street_id": "[]ADDRESS.RealAddressStreetId",
          "$$documents.1000014.ref_num": "COMPLEX_OMS.oms",
          "$$addresses.1000009.stroenie_no": "[]REG_ADDRESS.RegAddressBuilding",
          "$$addresses.7.moscow_district_id": "[]ADDRESS.RealAddressDistrictId",
          "$$addresses.7.unad": "[]ADDRESS.RealAddressUnad",
          "vehicles.pts_number": "[]VEHICLE.VehiclePTSNumber",
          "$$documents.1000016.identification_issuer": "PASSPORT_RF.PassportIssuer",
          "etalon_id": "USER.LoginName",
          "$$addresses.1000009.street_id": "[]REG_ADDRESS.RegAddressStreetId",
          "$$escredentials.1000003.etalon_id": "[]GIA.GIAId",
          "$$addresses.1000009.moscow_district_name": "[]REG_ADDRESS.RegAddressDistrict",
          "given_name_one": "FIO.FirstName",
          "vehicles.vin_number": "[]VEHICLE.VehicleVin",
          "$$escredentials.1000004.login": "[]OLYMPIAD.OlympiadLogin",
          "$$escredentials.1000003.doc_number": "[]GIA.GIADocNumber",
          "$$addresses.1000009.moscow_area_name": "[]REG_ADDRESS.RegAddressArea",
          "gkupays.gkupay_address_rel": "[]HOUSE.EpdAddressId",
          "$$addresses.7.building_name": "[]ADDRESS.RealAddressHouseId",
          "$$addresses.7.moscow_area_name": "[]ADDRESS.RealAddressArea",
          "$$addresses.7.moscow_district_name": "[]ADDRESS.RealAddressDistrict",
          "gender_tp_code": "PERSON.Gender",
          "$$addresses.7.stroenie_no": "[]ADDRESS.RealAddressBuilding",
          "$$documents.1000013.ref_num": "TAX_INFO.INN",
          "$$addresses.1000009.building_name": "[]REG_ADDRESS.RegAddressHouseId",
          "$$escredentials.1000004.description": "[]OLYMPIAD.OlympiadComment",
          "birth_dt": "PERSON.Birthday",
          "gkupays.epd": "[]HOUSE.EpdAddressNumber",
          "$$addresses.1000009.house_no": "[]REG_ADDRESS.RegAddressHouse",
          "$$addresses.1000009.street_omk": "[]REG_ADDRESS.RegAddressStreetOMK",
          "$$citizen_relatives.OTHERS.etalon_id": "[]RELATIVES.RltvId",
          "vehicles.stsnumber": "[]VEHICLE.VehicleSTSNumber",
          "vehicles.description": "[]VEHICLE.VehicleDescription",
          "$$addresses.1000009.unad": "[]REG_ADDRESS.RegAddressUnad",
          "$$escredentials.1000004.password": "[]OLYMPIAD.OlympiadPassword",
          "gkupays.phone": "[]HOUSE.EpdAddressCityPhone",
          "$$addresses.7.street_omk": "[]ADDRESS.RealAddressStreetOMK",
          "$$documents.1000019.start_dt": "DRIVER_LICENSE.DriverLicenseIssueDate",
          "$$addresses.1000009.corpus_no": "[]REG_ADDRESS.RegAddressCorpus",
          "$$citizen_relatives.1000004.etalon_id": "[]CHILDREN.ChildId",
          "$$documents.1000016.identification_issuer_code": "PASSPORT_RF.PassportIssuerCode",
          "$$escredentials.1000003.login": "[]GIA.GIARegCode",
          "$$addresses.1000009.residence_num": "[]REG_ADDRESS.RegAddressFlat",
          "given_name_two": "FIO.MiddleName",
          "$$contacts.2.ref_num": "REG_DATA.Email",
          "$$documents.1000015.ref_num": "SNILS.snils",
          "$$documents.1000016.ref_num": "PASSPORT_RF.PassportNumber",
          "$$addresses.7.street_name_ex": "[]ADDRESS.RealAddressStreet",
          "$$documents.1000014.etalon_id": "COMPLEX_OMS.OmsId",
          "$$escredentials.1000003.description": "[]GIA.GIAComment",
          "$$citizen_relatives.1000003.etalon_id": "[]CHILDREN.ChildId",
          "$$documents.1000016.identification_passport_birth_place": "PASSPORT_RF.PassportBirthPlace"
        }
      },
      "formattingRules": {
        "1": {
          "birth_dt": {
            "inputFormat": "2006-01-02 15:04:05.999",
            "type": "DATE_TIME",
            "outFormat": "2006-01-02"
          },
          "documents.1000016.start_dt": {
            "type": "DATE_TIME",
            "outFormat": "2006-01-02",
            "inputFormat": "2006-01-02 15:04:05.999"
          }
        }
      }
    }')::jsonb);

-- +goose Down

