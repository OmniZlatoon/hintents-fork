// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

//! Lightweight harness context for rollback-and-resume control commands.

use crate::types::SimulationRequest;

/// Command emitted by the Go bridge when replaying from a rewound point.
pub const ROLLBACK_AND_RESUME: &str = "ROLLBACK_AND_RESUME";

/// Tracks fork/replay metadata for one simulator process invocation.
#[derive(Debug, Default, Clone)]
pub struct HarnessContext {
    pub last_rewind_step: Option<u32>,
    pub fork_count: u32,
    pub reset_count: u32,
}

impl HarnessContext {
    /// Applies control-command side effects and returns human-readable log lines.
    pub fn apply_control_command(&mut self, request: &SimulationRequest) -> Vec<String> {
        let mut logs = Vec::new();

        let Some(command) = request.control_command.as_deref() else {
            return logs;
        };

        if command.eq_ignore_ascii_case(ROLLBACK_AND_RESUME) {
            if request.harness_reset {
                self.reset_temporary_state();
                logs.push("Harness temporary state reset before replay".to_string());
            }

            self.fork_count = self.fork_count.saturating_add(1);
            self.last_rewind_step = request.rewind_step;

            let mut replay_log = format!(
                "Bridge command {} accepted (rewind_step={})",
                ROLLBACK_AND_RESUME,
                request.rewind_step.unwrap_or(0)
            );
            if let Some(params) = &request.fork_params {
                replay_log.push_str(&format!(", fork_params={}", serde_json::to_string(params).unwrap_or_default()));
            }
            logs.push(replay_log);
            return logs;
        }

        logs.push(format!("Bridge command ignored: {}", command));
        logs
    }

    /// Resets temporary harness counters that should not leak across forks.
    pub fn reset_temporary_state(&mut self) {
        self.reset_count = self.reset_count.saturating_add(1);
    }
}
