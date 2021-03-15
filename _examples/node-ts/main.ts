
import * as c from "cowsay2";
import { argv } from "process";

console.log(c.say(argv[2] || "Hello, World!"));
