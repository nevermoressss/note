package studygo

import "sort"

func reconstructQueue(people [][]int) [][]int {
	sortS(people)
	var  res [][]int
	for _,v:=range people{
		temp:=append([][]int{v},res[v[1]:]...)
		res=append(res[:v[1]],temp...)
	}
	return res
}

type S [][]int

func sortS(s S){
	sort.Sort(s)
}

func (s S)Swap(i,j int) {
	s[i],s[j]=s[j],s[i]
}

func (s S)Len() int{
	return len(s)
}

func (s S)Less(i,j int) bool{
	if s[i][0]>s[j][0]==true{
		return true
	}
	if s[i][0]==s[j][0]{
		if s[i][1]<s[j][1]{
			return true
		}
		return false
	}
	return false
}