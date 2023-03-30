import { hash } from "starknet";

console.log(hash.getSelectorFromName("Ownable::owner"));

console.log(hash.getSelectorFromName("Ownable:proposedOwner"));