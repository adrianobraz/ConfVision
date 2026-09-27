#[derive(Debug, Clone, Copy)]
pub struct PersonDetection {
    pub class_id: i32,
    pub confidence: f64,
    pub xyxy: [f64; 4],
}
