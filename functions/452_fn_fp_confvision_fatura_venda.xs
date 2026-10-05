// Venda de licencas ConfVision: cria licenca pendente no Postgres + fatura aberta no Xano
function fn_fp_confvision_fatura_venda {
  input {
    text id_franqueado? filters=trim
    object[] itens? {
      schema {
        text plano? filters=trim
        int quantidade?=1 filters=min:1
      }
    }

    text origem?=admin filters=trim
    text admin_usuario? filters=trim
  }

  stack {
    precondition (($input.id_franqueado|is_empty) == false) {
      error = "id_franqueado obrigatorio"
    }

    precondition (($input.itens != null) && (($input.itens|count) > 0)) {
      error = "itens obrigatorio"
    }

    function.run fn_fp_franqueado_id_central {
      input = {id_franqueado: $input.id_franqueado}
    } as $escopo_fra

    var $id_rep {
      value = $escopo_fra.id_representante
    }

    var $id_central {
      value = $escopo_fra.id_central
    }

    var $valor_total {
      value = 0
    }

    var $lote_ops {
      value = []
    }

    foreach ($input.itens) {
      each as $item {
        precondition (($item.plano|is_empty) == false) {
          error = "plano obrigatorio em cada item"
        }

        function.run fn_vis_plano_flags {
          input = {plano: $item.plano}
        } as $flags

        precondition ($flags.plano_label != "Nenhum") {
          error = "Plano invalido: " ~ $item.plano
        }

        function.run fn_fp_confvision_valor_com_desconto {
          input = {
            id_franqueado: $input.id_franqueado
            valor_base   : $flags.valor
          }
        } as $preco

        var $qtd {
          value = $item.quantidade|first_notempty:1
        }

        var $desc_item {
          value = "ConfVision — " ~ $flags.plano_label
        }

        conditional {
          if ($preco.desconto_aplicado) {
            var.update $desc_item {
              value = $desc_item ~ " (20% Pro+)"
            }
          }
        }

        var $obs_lic {
          value = "Aguardando pagamento — " ~ $flags.plano_label
        }

        for (50) {
          each as $n {
            conditional {
              if ((($n + 1) > $qtd)) {
                break
              }
            }

            var $rec_ops {
              value = {
                id_franqueado: $input.id_franqueado
                plano        : $item.plano
                unidade      : $flags.unidade
                valor        : $preco.valor
                status       : "pendente"
                observacao   : $obs_lic
                desc_fatura  : $desc_item
              }
            }

            var.update $lote_ops {
              value = $lote_ops|push:$rec_ops
            }

            var.update $valor_total {
              value = $valor_total + $preco.valor
            }
          }
        }
      }
    }

    function.run fn_fp_confvision_postgres_sync_licenca {
      input = {
        acao    : "criar_lote"
        licencas: $lote_ops
      }
    } as $criacao

    precondition ($criacao.ok) {
      error = $criacao.erro|first_notempty:"falha_criar_licencas_postgres"
    }

    var $licencas_criadas {
      value = $criacao.resposta|get:"licencas":[]
    }

    precondition (($licencas_criadas|count) == ($lote_ops|count)) {
      error = "criacao_licencas_incompleta"
    }

    var $itens_fatura_final {
      value = []
    }

    foreach ($licencas_criadas) {
      each as $lic_row {
        var $desc_fat {
          value = $lic_row|get:"desc_fatura":null
        }

        conditional {
          if ($desc_fat == null || ($desc_fat|is_empty)) {
            function.run fn_vis_plano_flags {
              input = {plano: $lic_row|get:"plano":""}
            } as $flags_lic

            var.update $desc_fat {
              value = "ConfVision — " ~ $flags_lic.plano_label
            }
          }
        }

        var $lic_id_pg {
          value = $lic_row|get:"id":null|to_int
        }

        precondition ($lic_id_pg > 0) {
          error = "licenca_postgres_sem_id_na_resposta_ops"
        }

        var $item_fat {
          value = {
            descricao     : $desc_fat
            valor_unitario: $lic_row|get:"valor":0
            ref_id        : $lic_id_pg|to_text
          }
        }

        var.update $itens_fatura_final {
          value = $itens_fatura_final|push:$item_fat
        }
      }
    }

    var $vencimento {
      value = now|add_secs_to_timestamp:5 * 86400
    }

    var $referencia {
      value = "FAT-CV-" ~ (now|format_timestamp:"YmdHis":"UTC") ~ "-" ~ $input.id_franqueado
    }

    var $obs {
      value = "Venda ConfVision — aguardando pagamento"
    }

    conditional {
      if ($input.origem == "franqueado") {
        var.update $obs {
          value = "Compra ConfVision pelo portal — aguardando confirmacao da central"
        }
      }
    }

    db.add fp_fatura {
      data = {
        created_at      : "now"
        id_franqueado   : $input.id_franqueado
        id_representante: $id_rep
        id_central      : $id_central
        referencia      : $referencia
        status          : "aberta"
        tipo            : "confvision_venda"
        valor_total     : $valor_total
        vencimento_em   : $vencimento
        observacao      : $obs
      }
    } as $fatura

    var $itens_db {
      value = []
    }

    foreach ($itens_fatura_final) {
      each as $fi {
        db.add fp_fatura_item {
          data = {
            created_at    : "now"
            fp_fatura_id  : $fatura.id
            descricao     : $fi|get:"descricao":""
            quantidade    : 1
            valor_unitario: $fi|get:"valor_unitario":0
            valor_total   : $fi|get:"valor_unitario":0
            ref_tipo      : "vis_licenca"
            ref_id        : $fi|get:"ref_id":""
          }
        } as $item_db

        var.update $itens_db {
          value = $itens_db|push:$item_db
        }
      }
    }

    var $qtd_lic {
      value = $licencas_criadas|count
    }

    var $detalhe_log {
      value = $qtd_lic|to_text
    }

    var.update $detalhe_log {
      value = $detalhe_log ~ " licencas"
    }

    conditional {
      if ($input.origem == "franqueado") {
        var.update $detalhe_log {
          value = $detalhe_log ~ " (portal ConfVision)"
        }
      }
    }

    function.run fn_fp_financeiro_log {
      input = {
        acao            : "confvision_fatura_venda"
        id_franqueado   : $input.id_franqueado
        id_representante: $id_rep
        id_central      : $id_central
        ref_tipo        : "fp_fatura"
        ref_id          : $fatura.id|to_text
        detalhe         : $detalhe_log
        origem          : $input.origem|first_notempty:"admin"
        valor           : $valor_total
        produto         : "confvision"
        admin_usuario   : $input.admin_usuario
      }
    } as $log

    db.get fp_fatura {
      field_name = "id"
      field_value = $fatura.id
    } as $fatura_mirror

    var $mirror_entidades {
      value = []
    }

    array.push $mirror_entidades {
      value = {
        tipo : "fatura"
        acao : "upsert"
        dados: $fatura_mirror
      }
    }

    function.run fn_fp_fin_mirror_hook {
      runtime_mode = "async-shared"
      input = {entidades: $mirror_entidades}
    } as $mirror_hook
  }

  response = {
    fatura  : $fatura
    licencas: $licencas_criadas
    itens   : $itens_db
    log     : $log
  }
}
