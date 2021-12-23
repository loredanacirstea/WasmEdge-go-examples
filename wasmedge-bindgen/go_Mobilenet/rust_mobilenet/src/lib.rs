use std::time::Instant;
use wasmedge_tensorflow_interface;
use wasmedge_bindgen::*;
use wasmedge_bindgen_macro::*;

#[build_run]
fn infer(image_data: Vec<u8>) -> Result<Vec<u8>, String> {
    let start = Instant::now();
    let img = match image::load_from_memory(&image_data[..]) {
        Ok(a) => a.to_rgb8(),
        Err(e) => {
            println!("{:?}", e);
            panic!();
        }
    };
    println!("RUST: Loaded image in ... {:?}", start.elapsed());
    let resized = image::imageops::thumbnail(&img, 224, 224);
    println!("RUST: Resized image in ... {:?}", start.elapsed());
    let mut flat_img: Vec<f32> = Vec::new();
    for rgb in resized.pixels() {
        flat_img.push(rgb[0] as f32 / 255.);
        flat_img.push(rgb[1] as f32 / 255.);
        flat_img.push(rgb[2] as f32 / 255.);
    }

    let model_data: &[u8] = include_bytes!("mobilenet_v2_1.4_224_frozen.pb");
    let labels = include_str!("imagenet_slim_labels.txt");

    let mut session = wasmedge_tensorflow_interface::Session::new(
        model_data,
        wasmedge_tensorflow_interface::ModelType::TensorFlow,
    );
    session
        .add_input("input", &flat_img, &[1, 224, 224, 3])
        .add_output("MobilenetV2/Predictions/Softmax")
        .run();
    let res_vec: Vec<f32> = session.get_output("MobilenetV2/Predictions/Softmax");
    println!("RUST: Parsed output in ... {:?}", start.elapsed());

    let mut i = 0;
    let mut max_index: i32 = -1;
    let mut max_value: f32 = -1.0;
    while i < res_vec.len() {
        let cur = res_vec[i];
        if cur > max_value {
            max_value = cur;
            max_index = i as i32;
        }
        i += 1;
    }
    println!("RUST: index {}, prob {}", max_index, max_value);

    let mut confidence = "low";
    if max_value > 0.75 {
        confidence = "very high";
    } else if max_value > 0.5 {
        confidence = "high";
    } else if max_value > 0.2 {
        confidence = "medium";
    }

    let mut label_lines = labels.lines();
    for _i in 0..max_index {
        label_lines.next();
    }
    let ret: (String, String) = (
        label_lines.next().unwrap().to_string(),
        confidence.to_string(),
    );
    println!(
        "RUST: Finished post-processing in ... {:?}",
        start.elapsed()
    );
    let ret = serde_json::to_string(&ret).unwrap();
    Ok(ret.as_bytes().to_vec())
}