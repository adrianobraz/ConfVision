// Timeline de segmentos por camera
query vis_gravacao_segmento_by_camera verb=GET {
  api_group = "confVision"

  input {
    int vis_camera_id? filters=min:1
    timestamp? de?
    timestamp? ate?
    int page?=1
  }

  stack {
    db.query vis_gravacao_segmento {
      where = $db.vis_gravacao_segmento.vis_camera_id == $input.vis_camera_id && $db.vis_gravacao_segmento.inicio_em >=? $input.de && $db.vis_gravacao_segmento.inicio_em <=? $input.ate && $db.vis_gravacao_segmento.status == "disponivel"
      sort = {vis_gravacao_segmento.inicio_em: "desc"}
      return = {type: "list", paging: {page: $input.page, per_page: 50}}
    } as $resultado
  }

  response = {dados: $resultado}
}