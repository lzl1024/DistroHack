public class Test {
	public static void main(String[] args) {
		String[] tokens = new String[] { "2", "1", "+", "3", "*" };
	  	Source s = new Source();
	  	int result = s.evalRPN(tokens); 
	  	System.out.println(result);
	}
}