fun main(args: Array<String>) {
    java.util.Scanner(System.`in`).use { scan ->
        while (scan.hasNextLine()) println(scan.nextLine().uppercase())
    }
}
