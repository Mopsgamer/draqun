const std = @import("std");
const runner = @import("runner");
const native_sdk = @import("native_sdk");

pub const panic = std.debug.FullPanic(native_sdk.debug.capturePanic);

const server_url = "http://localhost:3000/";
const server_origin = "http://localhost:3000";

const DraqunApp = struct {
    fn app(self: *@This()) native_sdk.App {
        return .{
            .context = self,
            .name = "draqun",
            .source = native_sdk.WebViewSource.url(server_url),
        };
    }
};

pub fn main(init: std.process.Init) !void {
    var app_state = DraqunApp{};
    try runner.runWithOptions(app_state.app(), .{
        .app_name = "Draqun",
        .window_title = "Draqun",
        .bundle_id = "com.mopsgamer.draqun",
        .icon_path = "assets/icon.png",
        .default_frame = native_sdk.geometry.RectF.init(0, 0, 1200, 800),
        .security = .{
            .navigation = .{
                .allowed_origins = &.{server_origin},
            },
        },
    }, init);
}
