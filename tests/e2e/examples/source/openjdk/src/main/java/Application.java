import io.javalin.Javalin;

public class Application {

  public static void main(String[] args) {
    final Javalin app = Javalin.create().start(8080);
    app.get("/", ctx -> ctx.result("Hello World from Javalin!"));
  }

}
