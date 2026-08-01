table vis_camera {
  auth = false

  schema {
    int id
    timestamp created_at?=now {
      visibility = "private"
    }
  
    bool ativo?
    bool bloqueado?
    text nome? filters=trim
    text id_franqueado? filters=trim
    text id_cliente? filters=trim
    text id_dispositivo? filters=trim
    text conta? filters=trim
    text particao? filters=trim
    text canal? filters=trim
    text setor? filters=trim
    text protocolo? filters=trim
    text rtsp_url_sec? filters=trim
    text onvif_host? filters=trim
    int onvif_porta?
    text onvif_usuario? filters=trim
    text onvif_senha? filters=trim
    decimal confianca_min?
    int cooldown_seg?
    bool somente_armado?
    bool deteccao_humano?
    bool deteccao_veiculo?
    text status? filters=trim
    timestamp? ultimo_evento_em?
    text worker_id? filters=trim
    timestamp? ultimo_ping_em?
    text zonauser? filters=trim
    bool captura_sensor?
    bool captura_analitico?
    bool analitico_pausado?
    text id_setor? filters=trim
    text snapshot_url? filters=trim
    int vis_licenca_id? {
      table = "vis_licenca"
    }
  
    text plano? filters=trim
    timestamp? ativado_em?
    bool evento_grava_foto?
    bool evento_grava_video?
    int? vis_licenca_gravacao_id?
    bool grava_continua?
    int? retencao_dias?
    timestamp? gravacao_ativada_em?
    text gravacao_status? filters=trim
    bool grava_movimento?
    bool grava_timelapse?
    bool gravacao_flush_pedido?
    text modo_deteccao? filters=trim
    bool bloqueado?
    bool analitico_pausado?
    int vis_mediamtx_node_id? {
      table = "vis_mediamtx_node"
    }
  }

  index = [
    {type: "primary", field: [{name: "id"}]}
    {type: "gin", field: [{name: "xdo", op: "jsonb_path_op"}]}
    {type: "btree", field: [{name: "created_at", op: "desc"}]}
    {type: "btree", field: [{name: "id_franqueado", op: "asc"}]}
    {type: "btree", field: [{name: "id_cliente", op: "asc"}]}
    {type: "btree", field: [{name: "id_dispositivo", op: "asc"}]}
    {
      type : "btree"
      field: [
        {name: "id_dispositivo", op: "asc"}
        {name: "particao", op: "asc"}
        {name: "zonauser", op: "asc"}
      ]
    }
    {type: "btree", field: [{name: "worker_id", op: "asc"}]}
    {type: "btree", field: [{name: "vis_mediamtx_node_id", op: "asc"}]}
    {type: "btree", field: [{name: "vis_licenca_id", op: "asc"}]}
    {
      type : "btree"
      field: [{name: "vis_licenca_gravacao_id", op: "asc"}]
    }
    {
      type : "btree"
      field: [
        {name: "grava_continua", op: "asc"}
        {name: "id_franqueado", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "grava_movimento", op: "asc"}
        {name: "id_franqueado", op: "asc"}
      ]
    }
    {
      type : "btree"
      field: [
        {name: "grava_timelapse", op: "asc"}
        {name: "id_franqueado", op: "asc"}
      ]
    }
  ]
}