// Timeline de segmentos por franqueado (historico — inclui cameras inativas)
query vis_gravacao_segmento_by_franqueado verb=GET {
  api_group = "confVision"

  input {
    text id_franqueado? filters=trim
    timestamp? de?
    timestamp? ate?
    int page?=1
  }

  stack {
    db.query vis_gravacao_segmento {
      where = $db.vis_gravacao_segmento.id_franqueado == $input.id_franqueado && $db.vis_gravacao_segmento.inicio_em >=? $input.de && $db.vis_gravacao_segmento.inicio_em <=? $input.ate && $db.vis_gravacao_segmento.status == "disponivel"
      sort = {vis_gravacao_segmento.inicio_em: "desc"}
      return = {type: "list", paging: {page: $input.page, per_page: 50}}
    } as $resultado
  }

  response = {dados: $resultado}
}