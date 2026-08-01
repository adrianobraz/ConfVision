// Preenche snapshot modulos/limites na assinatura a partir do catalogo/plano (se vazio)
function fn_fp_assinatura_aplicar_catalogo {
  input {
    int assinatura_id? filters=min:1
  }

  stack {
    precondition ($input.assinatura_id != null) {
      error = "assinatura_id obrigatorio"
    }
  
    db.get fp_assinatura_produto {
      field_name = "id"
      field_value = $input.assinatura_id
    } as $assinatura
  
    precondition ($assinatura != null) {
      error = "Assinatura nao encontrada"
    }
  
    var $cat {
      value = null
    }
  
    conditional {
      if ($assinatura.fp_produto_catalogo_id != null && $assinatura.fp_produto_catalogo_id > 0) {
        db.get fp_produto_catalogo {
          field_name = "id"
          field_value = $assinatura.fp_produto_catalogo_id
        } as $cat
      }
    }
  
    conditional {
      if ($cat == null && ($assinatura.plano|is_empty) == false) {
        function.run fn_fp_catalogo_get_por_plano {
          input = {
            produto   : $assinatura.produto
            plano     : $assinatura.plano
            id_central: $assinatura.id_central|first_notempty:""
          }
        } as $cat
      }
    }
  
    function.run fn_fp_assinatura_regras_efetivas {
      input = {
        produto               : $assinatura.produto
        plano                 : $assinatura.plano
        tipo_contratacao      : $assinatura.tipo_contratacao
        modulos_snapshot      : $assinatura.modulos_json
        limites_snapshot      : $assinatura.limites_json
        fp_produto_catalogo_id: $assinatura.fp_produto_catalogo_id
      }
    } as $regras
  
    var $patch_modulos {
      value = $assinatura.modulos_json
    }
  
    var $patch_limites {
      value = $assinatura.limites_json
    }
  
    var $patch_catalogo_id {
      value = $assinatura.fp_produto_catalogo_id
    }
  
    conditional {
      if ($patch_modulos == null) {
        var.update $patch_modulos {
          value = $regras.modulos_json
        }
      }
    }
  
    conditional {
      if ($patch_limites == null) {
        var.update $patch_limites {
          value = $regras.limites_json
        }
      }
    }
  
    conditional {
      if (($patch_catalogo_id == null || $patch_catalogo_id <= 0) && $cat != null) {
        var.update $patch_catalogo_id {
          value = $cat.id
        }
      }
    }
  
    db.patch fp_assinatura_produto {
      field_name = "id"
      field_value = $input.assinatura_id
      data = {
        modulos_json          : $patch_modulos
        limites_json          : $patch_limites
        fp_produto_catalogo_id: $patch_catalogo_id
      }
    } as $model
  }

  response = $model
}