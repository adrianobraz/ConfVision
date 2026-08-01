query ListContactID verb=POST {
  api_group = "ConfmonitCentralAlarme"

  input {
  }

  stack {
    api.request {
      url = "http://185.130.61.4:2010/v4/contactid/listaPadrao"
      method = "POST"
    } as $api1
  }

  response = $api1
}