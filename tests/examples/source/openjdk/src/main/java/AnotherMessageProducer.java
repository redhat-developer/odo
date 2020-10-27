public final class AnotherMessageProducer {

    // this is a redundant class added for scenarios where we need to remove a file
    // and then build the component. As this file is not essential for a successful
    // build we can remove it for testing.

    private AnotherMessageProducer() {}

    public static String produce() {
      return "Hello World from Another Message Producer";
    }

}