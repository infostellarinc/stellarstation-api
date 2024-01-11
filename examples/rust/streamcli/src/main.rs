pub mod api {
    tonic::include_proto!("stellarstation.api.v1");
    pub mod radio {
        tonic::include_proto!("stellarstation.api.v1.radio");
    }
    pub mod antenna {
        tonic::include_proto!("stellarstation.api.v1.antenna");
    }
    pub mod monitoring {
        tonic::include_proto!("stellarstation.api.v1.monitoring");
    }
    pub mod orbit {
        tonic::include_proto!("stellarstation.api.v1.orbit");
    }
    pub mod common {
        tonic::include_proto!("stellarstation.api.v1.common");
    }
    pub mod groundstation {
        tonic::include_proto!("stellarstation.api.v1.groundstation");
    }
}

fn main() {
    println!("Hello, world!");
}
