// Lista usuarios ConfMonit por ID_Vinculo (Central / Representante / Franqueado)
// Central: id_vinculo = "CENTRAL". id_central da sessao pode ser catalogo ("CENTRAL")
// ou UUID real. A API Go filtra por IDCentralUUID — nunca enviar "CENTRAL" como UUID.
function fn_fp_usuario_listar_vinculo {
  input {
    text id_vinculo? filters=trim
    text id_central? filters=trim
  }

  stack {
    precondition (($input.id_vinculo|is_empty) == false) {
      error = "id_vinculo obrigatorio"
    }
  
    db.get fp_config_financeiro {
      field_name = "chave"
      field_value = "api_legada_url"
    } as $cfg_url
  
    var $api_base {
      value = "http://185.130.61.4:2010"
    }
  
    conditional {
      if ($cfg_url != null && ($cfg_url.valor|is_empty) == false) {
        var.update $api_base {
          value = $cfg_url.valor|trim
        }
      }
    }
  
    var $lista {
      value = []
    }
  
    var $api_status {
      value = ""
    }
  
    var $api_erro {
      value = ""
    }
  
    var $id_cen {
      value = $input.id_central|trim
    }
  
    var $uuid_filtro {
      value = ""
    }
  
    // UUID: com hifen OU 32+ chars hex legado. Nao aceitar "CENTRAL".
    conditional {
      if (($id_cen|is_empty) == false && $id_cen != "CENTRAL" && (($id_cen|contains:"-") || ($id_cen|strlen) >= 32)) {
        var.update $uuid_filtro {
          value = $id_cen
        }
      }
    }
  
    // Resolve catalogo → UUID via /v4/central/lista
    conditional {
      if (($uuid_filtro|is_empty) && (($id_cen|is_empty) == false || ($input.id_vinculo|trim|to_upper) == "CENTRAL")) {
        try_catch {
          try {
            api.request {
              url = $api_base ~ "/v4/central/lista"
              method = "POST"
              params = {}|set:"_":""
              headers = []
                |push:"Content-Type: application/json"
              timeout = 15
            } as $api_cen
          
            var $bc {
              value = $api_cen.response.result|first_notnull:{}
            }
          
            var $lc {
              value = $bc|get:"dados":($bc|get:"Dados":[])
            }
          
            var $cat_alvo {
              value = $id_cen
            }
          
            conditional {
              if (($cat_alvo|is_empty) || $cat_alvo == "CENTRAL") {
                var.update $cat_alvo {
                  value = "CENTRAL"
                }
              }
            }
          
            conditional {
              if ($lc != null && ($lc|is_array)) {
                foreach ($lc) {
                  each as $c {
                    var $id_cat {
                      value = $c
                        |get:"idCentral":($c|get:"ID_Central":"")
                        |trim
                    }
                  
                    var $uuid_c {
                      value = $c
                        |get:"idCentralUUID":($c|get:"IDCentralUUID":"")
                        |trim
                    }
                  
                    conditional {
                      if ($id_cat == $cat_alvo && (($uuid_c|is_empty) == false) && ($uuid_filtro|is_empty)) {
                        var.update $uuid_filtro {
                          value = $uuid_c
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        
          catch {
            var.update $uuid_filtro {
              value = $uuid_filtro
            }
          }
        }
      }
    }
  
    try_catch {
      try {
        conditional {
          if (($uuid_filtro|is_empty) == false) {
            api.request {
              url = $api_base ~ "/v4/usuario/listarByVinculo"
              method = "POST"
              params = {}
                |set:"idVinculo":$input.id_vinculo
                |set:"idCentralUUID":$uuid_filtro
              headers = []
                |push:"Content-Type: application/json"
              timeout = 20
            } as $api
          }

          else {
            api.request {
              url = $api_base ~ "/v4/usuario/listarByVinculo"
              method = "POST"
              params = {}|set:"idVinculo":$input.id_vinculo
              headers = []
                |push:"Content-Type: application/json"
              timeout = 20
            } as $api
          }
        }

        var $body {
          value = $api.response.result|first_notnull:{}
        }
      
        var.update $api_status {
          value = $body|get:"status":""
        }
      
        var $dados_api {
          value = $body|get:"dados":($body|get:"Dados":[])
        }
      
        conditional {
          if (($api_status|is_empty) == false && $api_status != "OK" && $api_status != "Vazio" && $api_status != "ok") {
            var.update $api_erro {
              value = $api_status
            }
          }
        }
      
        conditional {
          if ($dados_api != null && ($dados_api|is_array)) {
            foreach ($dados_api) {
              each as $u {
                var $id {
                  value = ($u|get:"idUsuario")
                    |first_notempty:($u|get:"ID_Usuario")
                    |first_notempty:""
                    |trim
                }

                var $nome_u {
                  value = ($u|get:"nome")
                    |first_notempty:($u|get:"Nome")
                    |first_notempty:""
                    |trim
                }

                var $nick_u {
                  value = ($u|get:"nick")
                    |first_notempty:($u|get:"Nick")
                    |first_notempty:""
                    |trim
                }

                var $email_u {
                  value = ($u|get:"email1")|first_notempty:""|trim
                }

                conditional {
                  if ($email_u|is_empty) {
                    var.update $email_u {
                      value = ($u|get:"Email1")|first_notempty:""|trim
                    }
                  }
                }

                conditional {
                  if ($email_u|is_empty) {
                    var.update $email_u {
                      value = ($u|get:"email")|first_notempty:""|trim
                    }
                  }
                }

                // Fallback: getDadosById quando listarByVinculo nao trouxe e-mail
                conditional {
                  if (($email_u|is_empty) && (($id|is_empty) == false)) {
                    try_catch {
                      try {
                        api.request {
                          url = $api_base ~ "/v4/usuario/getDadosById"
                          method = "POST"
                          params = {}|set:"idUsuario":$id
                          headers = []
                            |push:"Content-Type: application/json"
                          timeout = 10
                        } as $api_u

                        var $bu {
                          value = $api_u.response.result|first_notnull:{}
                        }

                        var $du {
                          value = $bu|get:"dados":($bu|get:"Dados":null)
                        }

                        conditional {
                          if ($du != null) {
                            var.update $email_u {
                              value = ($du|get:"email1")
                                |first_notempty:($du|get:"Email1")
                                |first_notempty:""
                                |trim
                            }

                            conditional {
                              if ($nome_u|is_empty) {
                                var.update $nome_u {
                                  value = ($du|get:"nome")|first_notempty:""|trim
                                }
                              }
                            }

                            conditional {
                              if ($nick_u|is_empty) {
                                var.update $nick_u {
                                  value = ($du|get:"nick")|first_notempty:""|trim
                                }
                              }
                            }
                          }
                        }
                      }

                      catch {
                        var.update $email_u {
                          value = $email_u
                        }
                      }
                    }
                  }
                }

                var $tel_u {
                  value = ($u|get:"telefone1")
                    |first_notempty:($u|get:"Telefone1")
                    |first_notempty:""
                    |trim
                }

                var $master_u {
                  value = ($u|get:"master")
                    |first_notempty:($u|get:"Master")
                    |first_notempty:"N"
                    |to_upper
                    |trim
                }

                var $adm_u {
                  value = ($u|get:"admFinanceiro")
                    |first_notempty:($u|get:"AdmFinanceiro")
                    |first_notempty:"N"
                    |to_upper
                    |trim
                }

                var $ativo_u {
                  value = ($u|get:"ativo")
                    |first_notempty:($u|get:"Ativo")
                    |first_notempty:"S"
                    |to_upper
                    |trim
                }

                var $tipo_u {
                  value = ($u|get:"tipo")|first_notempty:""|trim
                }

                var $id_vin_u {
                  value = ($u|get:"idVinculo")
                    |first_notempty:($u|get:"ID_Vinculo")
                    |first_notempty:$input.id_vinculo
                    |trim
                }

                conditional {
                  if (($id|is_empty) == false) {
                    var $item {
                      value = {}
                        |set:"id_usuario":$id
                        |set:"id_vinculo":$id_vin_u
                        |set:"nome":$nome_u
                        |set:"nick":$nick_u
                        |set:"email1":$email_u
                        |set:"email":$email_u
                        |set:"telefone1":$tel_u
                        |set:"master":$master_u
                        |set:"adm_financeiro":$adm_u
                        |set:"ativo":$ativo_u
                        |set:"tipo":$tipo_u
                    }
                  
                    array.push $lista {
                      value = $item
                    }
                  }
                }
              }
            }
          }
        }
      }
    
      catch {
        var.update $api_erro {
          value = $error.message|first_notempty:"falha ao listar usuarios"
        }

        var.update $lista {
          value = []
        }
      }
    }
  }

  response = {
    dados          : $lista
    total          : $lista|count
    id_vinculo     : $input.id_vinculo
    id_central     : $input.id_central
    id_central_uuid: $uuid_filtro
    api_status     : $api_status
    api_erro       : $api_erro
  }
}
