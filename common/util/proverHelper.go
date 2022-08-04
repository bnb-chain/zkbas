package util

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	COO_MODE = 1
	COM_MODE = 2
)

func LoadProvingKey(filepath string) (pk groth16.ProvingKey, err error) {
	fmt.Println("start reading proving key")
	pk = groth16.NewProvingKey(ecc.BN254)
	f, _ := os.Open(filepath)
	_, err = pk.ReadFrom(f)
	if err != nil {
		return pk, errors.New("read file error")
	}
	f.Close()

	return pk, nil
}

func LoadVerifyingKey(filepath string) (verifyingKey groth16.VerifyingKey, err error) {
	verifyingKey = groth16.NewVerifyingKey(ecc.BN254)
	f, _ := os.Open(filepath)
	_, err = verifyingKey.ReadFrom(f)
	if err != nil {
		return verifyingKey, errors.New("read file error")
	}
	f.Close()

	return verifyingKey, nil
}

func GenerateProof(
	r1cs frontend.CompiledConstraintSystem,
	provingKey groth16.ProvingKey,
	verifyingKey groth16.VerifyingKey,
	cBlock *cryptoBlock.Block,
) (proof groth16.Proof, err error) {
	// verify CryptoBlock
	blockWitness, err := cryptoBlock.SetBlockWitness(cBlock)
	if err != nil {
		logx.Errorf("[GenerateProof] unable to set block witness: %s", err.Error())
		return proof, err
	}
	var verifyWitness cryptoBlock.BlockConstraints
	verifyWitness.OldStateRoot = cBlock.OldStateRoot
	verifyWitness.NewStateRoot = cBlock.NewStateRoot
	verifyWitness.BlockCommitment = cBlock.BlockCommitment
	witness, err := frontend.NewWitness(&blockWitness, ecc.BN254)
	if err != nil {
		logx.Errorf("[GenerateProof] unable to generate new witness: %s", err.Error())
		return proof, err
	}
	vWitness, err := frontend.NewWitness(&verifyWitness, ecc.BN254, frontend.PublicOnly())
	if err != nil {
		logx.Errorf("[GenerateProof] unable to generate new witness: %s", err.Error())
		return proof, err
	}
	elapse := time.Now()
	logx.Info("start proving")
	proof, err = groth16.Prove(r1cs, provingKey, witness, backend.WithHints(std.Keccak256, std.ComputeSLp))
	if err != nil {
		logx.Errorf("[GenerateProof] unable to make a proof: %s", err.Error())
		return proof, err
	}
	fmt.Println("finish proving: ", time.Since(elapse))
	elapse = time.Now()
	logx.Info("start verifying")
	err = groth16.Verify(proof, verifyingKey, vWitness)
	if err != nil {
		logx.Errorf("[GenerateProof] invalid block proof: %s", err.Error())
		return proof, err
	}

	return proof, nil
}

func VerifyProof(
	proof groth16.Proof,
	vk groth16.VerifyingKey,
	cBlock *cryptoBlock.Block,
) error {
	// verify CryptoBlock
	blockWitness, err := cryptoBlock.SetBlockWitness(cBlock)
	if err != nil {
		logx.Errorf("[VerifyProof] unable to set block witness: %s", err.Error())
		return err
	}

	var verifyWitness cryptoBlock.BlockConstraints
	verifyWitness.OldStateRoot = cBlock.OldStateRoot
	verifyWitness.NewStateRoot = cBlock.NewStateRoot
	verifyWitness.BlockCommitment = cBlock.BlockCommitment
	_, err = frontend.NewWitness(&blockWitness, ecc.BN254)
	if err != nil {
		logx.Errorf("[VerifyProof] unable to generate new witness: %s", err.Error())
		return err
	}

	vWitness, err := frontend.NewWitness(&verifyWitness, ecc.BN254, frontend.PublicOnly())
	if err != nil {
		logx.Errorf("[VerifyProof] unable to generate new witness: %s", err.Error())
		return err
	}

	err = groth16.Verify(proof, vk, vWitness)
	if err != nil {
		logx.Errorf("[VerifyProof] invalid block proof: %s", err.Error())
		return err
	}
	return nil
}

type FormattedProof struct {
	A      [2]*big.Int
	B      [2][2]*big.Int
	C      [2]*big.Int
	Inputs [3]*big.Int
}

func FormatProof(oProof groth16.Proof, oldRoot, newRoot, commitment []byte) (proof *FormattedProof, err error) {
	proof = new(FormattedProof)
	const fpSize = 4 * 8
	var buf bytes.Buffer
	_, err = oProof.WriteRawTo(&buf)
	if err != nil {
		logx.Errorf("[FormatProof] unable to format proof: %s", err.Error())
		return nil, err
	}
	proofBytes := buf.Bytes()
	// proof.Ar, proof.Bs, proof.Krs
	proof.A[0] = new(big.Int).SetBytes(proofBytes[fpSize*0 : fpSize*1])
	proof.A[1] = new(big.Int).SetBytes(proofBytes[fpSize*1 : fpSize*2])
	proof.B[0][0] = new(big.Int).SetBytes(proofBytes[fpSize*2 : fpSize*3])
	proof.B[0][1] = new(big.Int).SetBytes(proofBytes[fpSize*3 : fpSize*4])
	proof.B[1][0] = new(big.Int).SetBytes(proofBytes[fpSize*4 : fpSize*5])
	proof.B[1][1] = new(big.Int).SetBytes(proofBytes[fpSize*5 : fpSize*6])
	proof.C[0] = new(big.Int).SetBytes(proofBytes[fpSize*6 : fpSize*7])
	proof.C[1] = new(big.Int).SetBytes(proofBytes[fpSize*7 : fpSize*8])

	// public witness
	proof.Inputs[0] = new(big.Int).SetBytes(oldRoot)
	proof.Inputs[1] = new(big.Int).SetBytes(newRoot)
	proof.Inputs[2] = new(big.Int).SetBytes(commitment)
	return proof, nil
}

func UnformatProof(proof *FormattedProof) (oProof groth16.Proof, err error) {
	var buf bytes.Buffer
	// write bytes to buffer
	buf.Write(proof.A[0].Bytes())
	buf.Write(proof.A[1].Bytes())
	buf.Write(proof.B[0][0].Bytes())
	buf.Write(proof.B[0][1].Bytes())
	buf.Write(proof.B[1][0].Bytes())
	buf.Write(proof.B[1][1].Bytes())
	buf.Write(proof.C[0].Bytes())
	buf.Write(proof.C[1].Bytes())

	// init oProof
	oProof = groth16.NewProof(ecc.BN254)

	// read buffer
	_, err = oProof.ReadFrom(bytes.NewReader(buf.Bytes()))
	if err != nil {
		logx.Errorf("[UnformatProof] unable to ReadFrom proof buffer: %s", err.Error())
		return oProof, err
	}

	return oProof, nil
}

func CompactProofs(proofs []*FormattedProof) []*big.Int {
	var res []*big.Int
	for _, proof := range proofs {
		res = append(res, proof.A[0])
		res = append(res, proof.A[1])
		res = append(res, proof.B[0][0])
		res = append(res, proof.B[0][1])
		res = append(res, proof.B[1][0])
		res = append(res, proof.B[1][1])
		res = append(res, proof.C[0])
		res = append(res, proof.C[1])
	}
	return res
}
