use std::{
    env, error::Error, ffi::OsStr, io::Write, os::unix::fs::PermissionsExt, path::Path,
    process::Command,
};

use clap::ValueEnum;
use rand::{distributions::Alphanumeric, Rng};

use crate::graph;
use crate::segments_scanner::SegmentsScanner;

#[derive(Clone, ValueEnum, PartialEq)]
pub enum Printer {
    Stdout,
    Executor,
    DesugarDebug,
}

impl Printer {
    pub fn print<I, S>(&self, targets: &[&str], content: &str, args: I) -> Result<(), Box<dyn Error>>
    where
        I: IntoIterator<Item = S>,
        S: AsRef<OsStr>,
    {
        let nodes: Vec<_> = SegmentsScanner::new(content).collect();
        let resolved = graph::resolve(&nodes, targets)
            .map_err(|err| -> Box<dyn Error> { format!("target not found: {}", err).into() })?;
        match self {
            Self::DesugarDebug => {
                println!("{}", content);
                Ok(())
            }
            Self::Stdout => {
                print!("{}", resolved);
                Ok(())
            }
            Self::Executor => {
                let prefix: String = rand::thread_rng()
                    .sample_iter(&Alphanumeric)
                    .take(5)
                    .map(char::from)
                    .collect();
                let path = Path::join(&env::temp_dir(), format!("mk-{}.tmp", prefix));
                let file_path = path.to_str().unwrap().to_string();
                {
                    let mut file = std::fs::File::create(&path)?;
                    let mut permissions = file.metadata()?.permissions();
                    permissions.set_mode(0o755);
                    file.set_permissions(permissions)?;
                    file.write_all(resolved.as_bytes())?;
                    file.flush()?;
                }
                Command::new(file_path).args(args).spawn()?.wait()?;
                std::fs::remove_file(path)?;
                Ok(())
            }
        }
    }
}
