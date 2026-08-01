// Delete WhatsEventCadBloq record
query "whatseventcadbloq/{whatseventcadbloq_id}" verb=DELETE {
  api_group = "franqueado"

  input {
    int whatseventcadbloq_id? filters=min:1
  }

  stack {
    db.del WhatsEventCadBloq {
      field_name = "id"
      field_value = $input.whatseventcadbloq_id
    }
  }

  response = null
}