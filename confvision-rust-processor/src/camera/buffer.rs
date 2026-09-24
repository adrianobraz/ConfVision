use std::collections::VecDeque;

/// Contador de frames com descarte de backlog (tempo real).
#[derive(Debug)]
pub struct BoundedFrameCounter {
    max_depth: usize,
    queue: VecDeque<u64>,
    pub dropped: u64,
    pub received: u64,
}

impl BoundedFrameCounter {
    pub fn new(max_depth: usize) -> Self {
        Self {
            max_depth: max_depth.max(1),
            queue: VecDeque::new(),
            dropped: 0,
            received: 0,
        }
    }

    pub fn push_frame(&mut self, seq: u64) {
        self.received += 1;
        if self.queue.len() >= self.max_depth {
            let _ = self.queue.pop_front();
            self.dropped += 1;
        }
        self.queue.push_back(seq);
    }

    pub fn depth(&self) -> usize {
        self.queue.len()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn drops_oldest_when_full() {
        let mut b = BoundedFrameCounter::new(2);
        b.push_frame(1);
        b.push_frame(2);
        b.push_frame(3);
        assert_eq!(b.received, 3);
        assert_eq!(b.dropped, 1);
        assert_eq!(b.depth(), 2);
    }
}
