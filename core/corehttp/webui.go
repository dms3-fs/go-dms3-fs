package corehttp

// TODO: move to DMS3NS
const WebUIPath = "/dms3fs/QmVSRy7pWP2zogUpCR5fqzemfaUGpocSBX2oAW11b7ms5z"

// this is a list of all past webUI paths.
var WebUIPaths = []string{
	WebUIPath,
// pre-fork webui not stored in dms3fs
//	"/dms3fs/QmXX7YRpU7nNBKfw75VG7Y1c3GwpSAGHRev67XVPgZFv9R",
//	"/dms3fs/QmXdu7HWdV6CUaUabd9q2ZeA4iHZLVyDRj3Gi4dsJsWjbr",
//	"/dms3fs/QmaaqrHyAQm7gALkRW8DcfGX3u8q9rWKnxEMmf7m9z515w",
//	"/dms3fs/QmSHDxWsMPuJQKWmVA1rB5a3NX2Eme5fPqNb63qwaqiqSp",
//	"/dms3fs/QmctngrQAt9fjpQUZr7Bx3BsXUcif52eZGTizWhvcShsjz",
//	"/dms3fs/QmS2HL9v5YeKgQkkWMvs1EMnFtUowTEdFfSSeMT4pos1e6",
//	"/dms3fs/QmR9MzChjp1MdFWik7NjEjqKQMzVmBkdK3dz14A6B5Cupm",
//	"/dms3fs/QmRyWyKWmphamkMRnJVjUTzSFSAAZowYP4rnbgnfMXC9Mr",
//	"/dms3fs/QmU3o9bvfenhTKhxUakbYrLDnZU7HezAVxPM6Ehjw9Xjqy",
//	"/dms3fs/QmPhnvn747LqwPYMJmQVorMaGbMSgA7mRRoyyZYz3DoZRQ",
}

var WebUIOption = RedirectOption("webui", WebUIPath)
