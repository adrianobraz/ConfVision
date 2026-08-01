// Update whatsappLigarErro record
query "whatsappligarerro/{whatsappligarerro_id}" verb=PUT {
  api_group = "task"

  input {
    int whatsappligarerro_id? filters=min:1
    dblink {
      table = "whatsappLigarErro"
    }
  }

  stack {
    util.get_raw_input {
      encoding = "json"
      exclude_middleware = false
    } as $raw_input
  
    db.patch whatsappLigarErro {
      field_name = "id"
      field_value = $input.whatsappligarerro_id
      data = `$input|pick:($raw_input|keys)`|filter_null
    } as $model
  }

  response = {dados: $model}
}